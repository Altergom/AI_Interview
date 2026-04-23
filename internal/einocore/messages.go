package einocore

import (
	"log/slog"

	"github.com/cloudwego/eino/schema"

	"ai_interview/internal/domain"
)

// MessagesFromSFT 将域内 SFT 行转为 Eino/LLM 使用的 schema.Message 切片（无业务状态，仅类型桥接）。
func MessagesFromSFT(msgs []domain.SFTMessage) []*schema.Message {
	out := make([]*schema.Message, 0, len(msgs))
	for _, m := range msgs {
		out = append(out, &schema.Message{
			Role:    sftRole(m.Role),
			Content: m.Content,
		})
	}
	return out
}

func sftRole(s string) schema.RoleType {
	switch s {
	case "system":
		return schema.System
	case "user":
		return schema.User
	case "assistant":
		return schema.Assistant
	case "tool":
		return schema.Tool
	default:
		slog.Warn("unknown SFT role, fallback to user", "role", s)
		return schema.User
	}
}
