// Package einocore 提供 Eino 框架的核心工具组件。
package einocore

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	einomodel "github.com/cloudwego/eino/components/model"
	"github.com/cloudwego/eino/schema"

	"ai_interview/internal/log"
	biz "ai_interview/internal/utils/respx"
)

const (
	// structuredOutputMaxRetries LLM JSON 解析失败最大重试次数。
	structuredOutputMaxRetries = 3
)

// StructuredOutputInvoker 封装「调用 LLM → 解析 JSON → 重试」的通用流程。
//
// 设计：
//   - 每次重试时将上一次的错误信息附加到 prompt，引导 LLM 修正输出
//   - 连续 maxRetries 次仍失败则返回 error，由调用方决定降级策略
//   - 不持有业务状态，可并发使用
type StructuredOutputInvoker struct {
	model      einomodel.BaseChatModel
	maxRetries int
}

// NewStructuredOutputInvoker 创建 StructuredOutputInvoker。
// maxRetries <= 0 时使用默认值 3。
func NewStructuredOutputInvoker(model einomodel.BaseChatModel, maxRetries int) *StructuredOutputInvoker {
	if maxRetries <= 0 {
		maxRetries = structuredOutputMaxRetries
	}
	return &StructuredOutputInvoker{
		model:      model,
		maxRetries: maxRetries,
	}
}

// Invoke 调用 LLM 并将响应解析为 JSON 写入 target（target 必须是指针）。
//
// 参数：
//   - systemPrompt: 系统提示词，描述任务和输出格式要求
//   - userContent: 用户消息，通常是待处理的原始文本
//   - target: 输出目标，必须是指向结构体的指针，如 *domain.StructuredResume
//
// 重试策略：JSON 解析失败时，在下一次请求中附加上一次的错误和 LLM 输出，
// 要求 LLM 修正。最多重试 maxRetries 次。
func (s *StructuredOutputInvoker) Invoke(ctx context.Context, systemPrompt, userContent string, target any) error {
	var (
		lastErr    error
		lastOutput string
	)

	for attempt := 1; attempt <= s.maxRetries; attempt++ {
		content := userContent
		// 从第二次起，把上次错误和输出附加进去，引导 LLM 自我修正
		if attempt > 1 && lastErr != nil {
			content = fmt.Sprintf(
				"%s\n\n---\n上一次输出（解析失败）：\n%s\n错误原因：%s\n\n请重新输出，只返回合法的 JSON，不要包含任何其他文字。",
				userContent, lastOutput, lastErr.Error(),
			)
		}

		msgs := []*schema.Message{
			{Role: schema.System, Content: systemPrompt},
			{Role: schema.User, Content: content},
		}

		resp, err := s.model.Generate(ctx, msgs)
		if err != nil {
			log.Warnf("[StructuredOutput] attempt %d/%d: LLM generate error: %v", attempt, s.maxRetries, err)
			lastErr = fmt.Errorf("LLM generate: %w", err)
			continue
		}

		raw := strings.TrimSpace(resp.Content)
		lastOutput = raw

		// 尝试从响应中提取 JSON（LLM 可能用 ```json ... ``` 包裹）
		raw = extractJSON(raw)

		if err := json.Unmarshal([]byte(raw), target); err != nil {
			log.Warnf("[StructuredOutput] attempt %d/%d: JSON unmarshal error: %v, raw=%q",
				attempt, s.maxRetries, err, truncate(raw, 200))
			lastErr = fmt.Errorf("JSON unmarshal: %w", err)
			continue
		}

		// 解析成功
		log.Infof("[StructuredOutput] succeeded on attempt %d/%d", attempt, s.maxRetries)
		return nil
	}

	return biz.WrapMsg(biz.CodeAIStructuredOutputFailed,
		biz.CodeAIStructuredOutputFailed.Message(),
		fmt.Errorf("all %d attempts failed, last: %w", s.maxRetries, lastErr),
	)
}

// extractJSON 从 LLM 输出中提取 JSON 内容。
// 处理常见格式：
//   - 原始 JSON
//   - ```json\n...\n``` 代码块
//   - ``` 代码块（无语言标记）
func extractJSON(s string) string {
	// 尝试去掉 markdown 代码块
	if idx := strings.Index(s, "```json"); idx >= 0 {
		s = s[idx+7:]
		if end := strings.LastIndex(s, "```"); end >= 0 {
			s = s[:end]
		}
		return strings.TrimSpace(s)
	}
	if idx := strings.Index(s, "```"); idx >= 0 {
		s = s[idx+3:]
		if end := strings.LastIndex(s, "```"); end >= 0 {
			s = s[:end]
		}
		return strings.TrimSpace(s)
	}
	// 找第一个 { 和最后一个 }
	start := strings.Index(s, "{")
	end := strings.LastIndex(s, "}")
	if start >= 0 && end > start {
		return s[start : end+1]
	}
	return s
}

// truncate 截断字符串，用于日志输出。
func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "..."
}
