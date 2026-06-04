package wiki

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"ai_interview/internal/einocore"
	"ai_interview/internal/llm"
	"ai_interview/internal/log"
)

// IngestResult 包含 ingest 的完整产物。
type IngestResult struct {
	Page    QuestionPage
	RawFile string
}

const ingestSystemPrompt = `你是技术面试出题助手。根据提供的原始技术文档，生成一道结构化面试题。

规则：
1. 提取文档的核心知识点，设计一个清晰的面试题
2. 考察点聚焦「面试官想通过这道题考察什么」
3. 答案要点 3-5 条，每条简洁扼要
4. 如果文档提到关联概念，生成 1-3 个追问链接，关系类型从以下三选一：
   - 深挖：有知识依赖，必须先懂当前题才能答下一题
   - 横向：同领域对比或替代方案
   - 跨域：跨技术领域的关联
5. slug 用英文小写，- 连接（如 "tcp-handshake"）
6. tags 从词汇表选：计算机网络、操作系统、数据库、缓存、Go、Java、大模型、中间件；难度从 易/中/高 选
7. 输出纯 JSON，不要 markdown 包裹`

// Ingest 执行单篇文档的 ingest 流程。
func Ingest(ctx context.Context, wikiDir, rawPath string) (*IngestResult, error) {
	relPath := strings.TrimPrefix(rawPath, filepath.Join(wikiDir, "raw")+string(os.PathSeparator))
	relPath = strings.ReplaceAll(relPath, "\\", "/")

	rawContent, err := os.ReadFile(rawPath)
	if err != nil {
		return nil, fmt.Errorf("[Wiki] read raw %s: %w", relPath, err)
	}

	// 构建 user content：schema 规则 + raw 文档
	userContent := fmt.Sprintf("原始文档路径：%s\n\n---\n\n%s", relPath, string(rawContent))

	page, err := generateQuestionPage(ctx, userContent)
	if err != nil {
		return nil, fmt.Errorf("[Wiki] generate page for %s: %w", relPath, err)
	}

	log.Infof("[Wiki] generated page slug=%s question=%s", page.Slug, page.Question[:min(50, len(page.Question))])

	// 写 questions/{slug}.md
	questionDir := filepath.Join(wikiDir, "questions")
	if err := os.MkdirAll(questionDir, 0755); err != nil {
		return nil, fmt.Errorf("[Wiki] mkdir questions: %w", err)
	}

	mdContent := page.renderMarkdown(fmt.Sprintf("raw/%s", relPath))
	questionPath := filepath.Join(questionDir, page.Slug+".md")
	if err := os.WriteFile(questionPath, []byte(mdContent), 0644); err != nil {
		return nil, fmt.Errorf("[Wiki] write question page: %w", err)
	}

	// 更新 index
	idx, err := readIndex(wikiDir)
	if err != nil {
		return nil, fmt.Errorf("[Wiki] read index: %w", err)
	}
	idx.addNode(page.Slug)
	idx.addEdges(page.Slug, *page)
	if err := idx.writeIndex(wikiDir); err != nil {
		return nil, fmt.Errorf("[Wiki] write index: %w", err)
	}

	// 追加 log
	if err := appendLog(wikiDir, relPath, page.Slug); err != nil {
		return nil, fmt.Errorf("[Wiki] append log: %w", err)
	}

	// 执行 lint，记录问题但不阻塞
	issues, err := lint(wikiDir, idx)
	if err != nil {
		log.Warnf("[Wiki] lint error: %v", err)
	}
	for _, issue := range issues {
		log.Warnf("[Wiki] lint %s: %s — %s", issue.Type, issue.Slug, issue.Detail)
	}

	return &IngestResult{
		Page:    *page,
		RawFile: relPath,
	}, nil
}

func generateQuestionPage(ctx context.Context, userContent string) (*QuestionPage, error) {
	chatModel, err := llm.Registry.NewChatModel(ctx, llm.RoleSelector)
	if err != nil {
		return nil, fmt.Errorf("create chat model: %w", err)
	}

	invoker := einocore.NewStructuredOutputInvoker(chatModel, 3)
	var page QuestionPage
	if err := invoker.Invoke(ctx, ingestSystemPrompt, userContent, &page); err != nil {
		return nil, fmt.Errorf("generate: %w", err)
	}
	return &page, nil
}

func extractJSON(s string) string {
	start := strings.Index(s, "{")
	end := strings.LastIndex(s, "}")
	if start >= 0 && end > start {
		return s[start : end+1]
	}
	return s
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
