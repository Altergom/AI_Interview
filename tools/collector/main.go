// tools/collector/main.go 小林coding知识库采集器。
//
//	用法：go run tools/collector/main.go [flags]
//	依赖：PG / RabbitMQ 已启动（docker compose up -d）
package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/joho/godotenv"

	"ai_interview/internal/config"
	"ai_interview/internal/domain"
	"ai_interview/internal/llm"
	"ai_interview/internal/log"
	"ai_interview/internal/mq"
	"ai_interview/internal/mq/mqclient"
	"ai_interview/internal/storage/postgres"
	"ai_interview/internal/wiki"
)

var (
	repoURL    = flag.String("repo-url", "https://github.com/CS-Learnings/xiaolincoder_CS-Base", "CS-Base GitHub repo URL")
	repoBranch = flag.String("repo-branch", "main", "Git branch")
	localPath  = flag.String("local-path", "", "跳过 clone，使用本地路径")
	dryRun     = flag.Bool("dry-run", false, "仅解析输出，不写 DB 不投 MQ")
	limit      = flag.Int("limit", 0, "最大导入数量（0=不限制）")
	category   = flag.String("category", "", "仅导入指定目录（如 网络；空=全部）")
)

func main() {
	flag.Parse()
	_ = godotenv.Load(".env", ".env.local")
	if err := config.Load(); err != nil {
		fmt.Fprintf(os.Stderr, "load config: %v\n", err)
		os.Exit(1)
	}

	log.Init(config.Cfg.LogLevel, config.Cfg.LogFormat, config.Cfg.Env)
	llm.Init(config.Cfg)

	ctx := context.Background()

	wikiDir := filepath.Join("internal", "wiki")

	repoPath, err := resolveRepo(ctx, wikiDir, *localPath, *repoURL, *repoBranch)
	if err != nil {
		fmt.Fprintf(os.Stderr, "resolve repo: %v\n", err)
		os.Exit(1)
	}

	mdFiles, err := collectFiles(repoPath, *category)
	if err != nil {
		fmt.Fprintf(os.Stderr, "collect files: %v\n", err)
		os.Exit(1)
	}

	log.Infof("[Collector] found %d markdown files", len(mdFiles))

	if *limit > 0 && *limit < len(mdFiles) {
		mdFiles = mdFiles[:*limit]
	}

	// 连接 PG 和 MQ
	var (
		repo  *postgres.BankQuestionRepo
		mqCli *mqclient.Client
	)
	if !*dryRun {
		repo, mqCli, err = initInfra(ctx)
		if err != nil {
			fmt.Fprintf(os.Stderr, "init infra: %v\n", err)
			os.Exit(1)
		}
		defer mqCli.Close()
	}

	stats := &Stats{}
	for _, f := range mdFiles {
		if err := processFile(ctx, wikiDir, f, repo, mqCli, *dryRun, stats); err != nil {
			log.Errorf("[Collector] process %s: %v", f, err)
			stats.Failed++
			continue
		}
	}

	stats.Print()
}

func resolveRepo(ctx context.Context, wikiDir, local, url, branch string) (string, error) {
	rawDir := filepath.Join(wikiDir, "raw")

	if local != "" {
		return local, nil
	}

	repoPath := filepath.Join(rawDir, "CS-Base")
	if _, err := os.Stat(repoPath); os.IsNotExist(err) {
		log.Infof("[Collector] cloning %s...", url)
		cmd := exec.CommandContext(ctx, "git", "clone", "--depth=1", "--branch="+branch, url, repoPath)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		if err := cmd.Run(); err != nil {
			return "", fmt.Errorf("git clone: %w", err)
		}
	}
	return repoPath, nil
}

func collectFiles(root, category string) ([]string, error) {
	var files []string
	err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() || filepath.Ext(path) != ".md" {
			return nil
		}
		if strings.HasPrefix(filepath.Base(path), "README") {
			return nil
		}
		rel, _ := filepath.Rel(root, path)
		if category != "" {
			parts := strings.SplitN(rel, string(os.PathSeparator), 2)
			if len(parts) < 2 || parts[0] != category {
				return nil
			}
		}
		files = append(files, path)
		return nil
	})
	return files, err
}

func initInfra(ctx context.Context) (*postgres.BankQuestionRepo, *mqclient.Client, error) {
	db, err := postgres.New(ctx, postgres.Options{
		DSN:             config.Cfg.PostgresDSN,
		MaxOpenConns:    config.Cfg.PGMaxOpenConns,
		MaxIdleConns:    config.Cfg.PGMaxIdleConns,
		ConnMaxLifetime: config.Cfg.PGConnMaxLifetime,
		ConnMaxIdleTime: config.Cfg.PGConnMaxIdleTime,
	})
	if err != nil {
		return nil, nil, fmt.Errorf("connect pg: %w", err)
	}

	repo := postgres.NewBankQuestionRepo(db.Gorm())

	mqCli, err := mqclient.New(config.Cfg.MQBrokerURL)
	if err != nil {
		db.Close()
		return nil, nil, fmt.Errorf("connect mq: %w", err)
	}

	if err := mqCli.DeclareQueue(mq.TopicVectorizeTask); err != nil {
		db.Close()
		mqCli.Close()
		return nil, nil, fmt.Errorf("declare queue: %w", err)
	}

	return repo, mqCli, nil
}

func processFile(ctx context.Context, wikiDir, path string, repo *postgres.BankQuestionRepo, mqCli *mqclient.Client, dry bool, stats *Stats) error {
	stats.Total++

	if dry {
		log.Infof("[Collector] [DRY-RUN] would ingest: %s", path)
		stats.Imported++
		return nil
	}

	result, err := wiki.Ingest(ctx, wikiDir, path)
	if err != nil {
		return err
	}

	rec := toBankQuestionRecord(result)

	exists, err := repo.ExistsByQuestion(ctx, rec.Question)
	if err != nil {
		return fmt.Errorf("check exists: %w", err)
	}
	if exists {
		log.Infof("[Collector] skip existing: %s", rec.Question[:minStr(50, len(rec.Question))])
		stats.Skipped++
		return nil
	}

	id, err := repo.Insert(ctx, rec)
	if err != nil {
		return fmt.Errorf("insert bank_questions: %w", err)
	}

	task := mq.VectorizeTask{QuestionID: id}
	if err := mqCli.Publish(ctx, mq.TopicVectorizeTask, task); err != nil {
		log.Errorf("[Collector] publish vectorize_task id=%s: %v", id, err)
	}

	log.Infof("[Collector] imported: %s → %s (id=%s)", filepath.Base(path), result.Page.Slug, id)
	stats.Imported++
	return nil
}

func toBankQuestionRecord(r *wiki.IngestResult) *domain.BankQuestionRecord {
	tags := mapWikiTags(r.Page.Tags)
	related := make([]string, 0, len(r.Page.FollowUp))
	for _, link := range r.Page.FollowUp {
		related = append(related, link.Slug)
	}

	answer := fmt.Sprintf("考察点：%s\n\n答案要点：\n%s",
		r.Page.FocusPoints,
		strings.Join(r.Page.AnswerPoints, "\n"),
	)

	return &domain.BankQuestionRecord{
		Question:            r.Page.Question,
		StandardAnswer:      answer,
		Tags:                tags,
		RelatedConcepts:     related,
		FollowupQuestionIDs: []string{},
		Difficulty:          mapDifficulty(r.Page.Difficulty),
	}
}

var wikiTagMap = map[string]string{
	"计算机网络": "network",
	"操作系统":   "os",
	"数据库":    "database",
	"缓存":     "cache",
	"go":     "golang",
	"java":   "java",
	"大模型":   "ai-agent",
	"中间件":   "middleware",
	"前端":    "frontend",
}

func mapWikiTags(wikiTags []string) []string {
	var out []string
	seen := make(map[string]bool)
	for _, t := range wikiTags {
		t = strings.TrimPrefix(t, "#")
		t = strings.TrimSpace(t)
		if mapped, ok := wikiTagMap[t]; ok && !seen[mapped] {
			out = append(out, mapped)
			seen[mapped] = true
		} else if !seen[t] {
			out = append(out, t)
			seen[t] = true
		}
	}
	if len(out) == 0 {
		out = []string{"general"}
	}
	return out
}

func mapDifficulty(d string) domain.Difficulty {
	switch strings.TrimSpace(d) {
	case "易", "低", "easy":
		return domain.DifficultyEasy
	case "高", "难", "hard":
		return domain.DifficultyHard
	default:
		return domain.DifficultyMedium
	}
}

// Stats 导入统计。
type Stats struct {
	Total    int
	Imported int
	Skipped  int
	Failed   int
}

func (s *Stats) Print() {
	log.Infof("[Collector] done: total=%d imported=%d skipped=%d failed=%d", s.Total, s.Imported, s.Skipped, s.Failed)
}

func minStr(a, b int) int {
	if a < b {
		return a
	}
	return b
}
