package agent

import (
	"context"
	"crypto/sha256"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/cloudwego/eino/adk"
	einofilesystem "github.com/cloudwego/eino/adk/filesystem"
	einoskill "github.com/cloudwego/eino/adk/middlewares/skill"
	"github.com/redis/go-redis/v9"

	"ai_interview/internal/llm"
	"ai_interview/internal/log"
	rediskeys "ai_interview/internal/storage/redis"
)

// SelectorConfig question_selector 的构造参数。
type SelectorConfig struct {
	// SkillsDir SKILL.md 所在的父目录，支持相对路径（相对于工作目录）和绝对路径。
	// 默认值 "internal/einocore/skills"。
	SkillsDir string
	// RedisClient 用于历史题目去重
	RedisClient *redis.Client
	// AskedQTTL 去重 Set 的 TTL，默认 24h（与面试状态 TTL 对齐）
	AskedQTTL time.Duration
}

// NewSelector 创建注入了 Skill middleware 的 question_selector Agent。
func NewSelector(ctx context.Context, cfg SelectorConfig) (*adk.ChatModelAgent, error) {
	if cfg.RedisClient == nil {
		return nil, fmt.Errorf("[selector] RedisClient is required")
	}
	if cfg.AskedQTTL == 0 {
		cfg.AskedQTTL = 24 * time.Hour
	}

	// 解析 SkillsDir 为绝对路径
	skillsDir, err := resolveSkillsDir(cfg.SkillsDir)
	if err != nil {
		return nil, fmt.Errorf("[selector] resolve skills dir: %w", err)
	}

	// 启动时校验：扫描 skills 目录确保至少有一个可用 Skill
	if err := validateSkillsDir(skillsDir); err != nil {
		return nil, fmt.Errorf("[selector] validate skills dir: %w", err)
	}

	// 构建 LLM
	model, err := llm.Registry.NewChatModel(ctx, llm.RoleSelector)
	if err != nil {
		return nil, fmt.Errorf("[selector] new chat model: %w", err)
	}

	// 构建 Skill middleware
	skillMiddleware, err := buildSkillMiddleware(ctx, skillsDir)
	if err != nil {
		return nil, fmt.Errorf("[selector] build skill middleware: %w", err)
	}

	// 构建 Agent
	agent, err := adk.NewChatModelAgent(ctx, &adk.ChatModelAgentConfig{
		Name:        "question_selector",
		Description: "面试出题 Agent，根据候选人方向加载对应 Skill，生成技术面试问题",
		Instruction: selectorInstruction,
		Model:       model,
		Handlers:    []adk.ChatModelAgentMiddleware{skillMiddleware},
	})
	if err != nil {
		return nil, fmt.Errorf("[selector] new chat model agent: %w", err)
	}

	log.Infof("[Selector] initialized, skills_dir=%s", skillsDir)
	return agent, nil
}

// selectorInstruction 是注入给 question_selector 的基础 Instruction。
// Skill middleware 会在此基础上追加 Skills System 说明和可用 Skill 列表。
// stage 参数让同一个 selector 在 intro/questioning/algorithm 中保持不同语义。
const selectorInstruction = `你是一位技术面试出题 Agent。

你的职责：
1. 根据输入的 stage 和岗位方向，调用对应的 skill 工具获取出题规则
2. 严格按照 stage 语义生成当前阶段需要的问题
3. 确保问题不与已出过的题目重复（已出题目会在 context 中提供）

工作流程：
1. 接收 stage: intro | questioning | algorithm | closing
2. 接收岗位方向（如 go-backend / java-backend / frontend / algorithm / ai-agent）
3. 调用 skill 工具加载该方向的完整出题规则
4. 按 stage 生成合适问题：
   - intro：生成背景、项目、技术栈追问
   - questioning：生成技术追问或下一道技术问题
   - algorithm：选择算法题
   - closing：通常不需要出题，除非用于澄清候选人反问

输出格式要求：
- 只返回问题本身，不要加序号、解析或参考答案
- 不要说"好的，这是问题："等前缀
- 直接输出问题文本`

// ------- 路径处理 -------

// resolveSkillsDir 将相对路径解析为绝对路径（相对于工作目录）。
func resolveSkillsDir(dir string) (string, error) {
	if dir == "" {
		dir = "internal/einocore/skills"
	}
	if filepath.IsAbs(dir) {
		return dir, nil
	}
	wd, err := os.Getwd()
	if err != nil {
		return "", fmt.Errorf("get working dir: %w", err)
	}
	return filepath.Join(wd, dir), nil
}

// validateSkillsDir 校验 skills 目录存在且至少有一个 SKILL.md。
func validateSkillsDir(dir string) error {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return fmt.Errorf("read skills dir %q: %w", dir, err)
	}
	count := 0
	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		skillFile := filepath.Join(dir, e.Name(), "SKILL.md")
		if _, err := os.Stat(skillFile); err == nil {
			count++
		}
	}
	if count == 0 {
		return fmt.Errorf("no SKILL.md found under %q", dir)
	}
	log.Infof("[Selector] found %d skill(s) in %s", count, dir)
	return nil
}

// ------- Skill middleware -------

// buildSkillMiddleware 用本地文件系统 backend 构建 Skill middleware。
func buildSkillMiddleware(ctx context.Context, skillsDir string) (adk.ChatModelAgentMiddleware, error) {
	skillBackend, err := einoskill.NewBackendFromFilesystem(ctx, &einoskill.BackendFromFilesystemConfig{
		Backend: &localFSBackend{},
		BaseDir: skillsDir,
	})
	if err != nil {
		return nil, fmt.Errorf("new skill backend: %w", err)
	}

	middleware, err := einoskill.NewMiddleware(ctx, &einoskill.Config{
		Backend: skillBackend,
	})
	if err != nil {
		return nil, fmt.Errorf("new skill middleware: %w", err)
	}

	return middleware, nil
}

// ------- 本地文件系统 Backend -------
// Eino 只内置了 InMemoryBackend，这里实现读本地磁盘的版本。

type localFSBackend struct{}

func (b *localFSBackend) LsInfo(_ context.Context, req *einofilesystem.LsInfoRequest) ([]einofilesystem.FileInfo, error) {
	entries, err := os.ReadDir(req.Path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}
	infos := make([]einofilesystem.FileInfo, 0, len(entries))
	for _, e := range entries {
		fi, err := e.Info()
		if err != nil {
			continue
		}
		infos = append(infos, einofilesystem.FileInfo{
			Path:       filepath.Join(req.Path, e.Name()),
			IsDir:      e.IsDir(),
			Size:       fi.Size(),
			ModifiedAt: fi.ModTime().UTC().Format(time.RFC3339),
		})
	}
	return infos, nil
}

func (b *localFSBackend) Read(_ context.Context, req *einofilesystem.ReadRequest) (*einofilesystem.FileContent, error) {
	data, err := os.ReadFile(req.FilePath)
	if err != nil {
		return nil, err
	}
	content := string(data)
	if req.Offset > 1 || req.Limit > 0 {
		lines := strings.Split(content, "\n")
		start := req.Offset - 1
		if start < 0 {
			start = 0
		}
		if start >= len(lines) {
			return &einofilesystem.FileContent{Content: ""}, nil
		}
		lines = lines[start:]
		if req.Limit > 0 && req.Limit < len(lines) {
			lines = lines[:req.Limit]
		}
		content = strings.Join(lines, "\n")
	}
	return &einofilesystem.FileContent{Content: content}, nil
}

func (b *localFSBackend) GrepRaw(_ context.Context, _ *einofilesystem.GrepRequest) ([]einofilesystem.GrepMatch, error) {
	return nil, nil // Skill middleware 不使用 Grep
}

func (b *localFSBackend) GlobInfo(_ context.Context, req *einofilesystem.GlobInfoRequest) ([]einofilesystem.FileInfo, error) {
	pattern := filepath.Join(req.Path, req.Pattern)
	matches, err := filepath.Glob(pattern)
	if err != nil {
		return nil, err
	}
	infos := make([]einofilesystem.FileInfo, 0, len(matches))
	for _, m := range matches {
		fi, err := os.Stat(m)
		if err != nil {
			continue
		}
		infos = append(infos, einofilesystem.FileInfo{
			Path:       m,
			IsDir:      fi.IsDir(),
			Size:       fi.Size(),
			ModifiedAt: fi.ModTime().UTC().Format(time.RFC3339),
		})
	}
	return infos, nil
}

func (b *localFSBackend) Write(_ context.Context, req *einofilesystem.WriteRequest) error {
	if err := os.MkdirAll(filepath.Dir(req.FilePath), 0o755); err != nil {
		return err
	}
	return os.WriteFile(req.FilePath, []byte(req.Content), 0o644)
}

func (b *localFSBackend) Edit(_ context.Context, req *einofilesystem.EditRequest) error {
	data, err := os.ReadFile(req.FilePath)
	if err != nil {
		return err
	}
	content := string(data)
	if req.ReplaceAll {
		content = strings.ReplaceAll(content, req.OldString, req.NewString)
	} else {
		idx := strings.Index(content, req.OldString)
		if idx == -1 {
			return fmt.Errorf("old string not found in %s", req.FilePath)
		}
		content = content[:idx] + req.NewString + content[idx+len(req.OldString):]
	}
	return os.WriteFile(req.FilePath, []byte(content), 0o644)
}

// ------- 历史题目去重 -------

// questionHash 计算题目文本的去重 hash（SHA-256 前 16 字节 hex）。
func questionHash(question string) string {
	sum := sha256.Sum256([]byte(question))
	return fmt.Sprintf("%x", sum[:16])
}

// MarkQuestionAsked 将题目 hash 写入 Redis Set，TTL 与面试状态对齐。
func MarkQuestionAsked(ctx context.Context, rdb *redis.Client, interviewID, question string, ttl time.Duration) error {
	key := rediskeys.InterviewAskedQuestionsKey(interviewID)
	hash := questionHash(question)
	pipe := rdb.Pipeline()
	pipe.SAdd(ctx, key, hash)
	pipe.Expire(ctx, key, ttl)
	_, err := pipe.Exec(ctx)
	if err != nil {
		return fmt.Errorf("[selector] mark asked question: %w", err)
	}
	return nil
}

// IsQuestionAsked 检查题目是否已出过。
func IsQuestionAsked(ctx context.Context, rdb *redis.Client, interviewID, question string) (bool, error) {
	key := rediskeys.InterviewAskedQuestionsKey(interviewID)
	hash := questionHash(question)
	exists, err := rdb.SIsMember(ctx, key, hash).Result()
	if err != nil {
		return false, fmt.Errorf("[selector] check asked question: %w", err)
	}
	return exists, nil
}
