package channel

import (
	"context"

	"gateway/inbound"
)

// ChannelAdapter 渠道适配器接口。
// 每个平台实现此接口，负责：签名验证、原始 payload 解析、转换为 InboundEvent。
type ChannelAdapter interface {
	// Name 返回渠道标识，如 "wechat" / "feishu" / "qqbot"。
	Name() string

	// VerifySignature 验证平台推送的签名，失败返回 error。
	VerifySignature(ctx context.Context, headers map[string]string, body []byte) error

	// Parse 将平台原始 payload 解析为统一入站事件。
	Parse(ctx context.Context, body []byte) (*inbound.InboundEvent, error)

	// Ack 返回平台要求的 ack 响应体（challenge 验证、success 字符串等）。
	Ack(ctx context.Context, body []byte) (any, error)

	// Send 向平台发送消息。
	Send(ctx context.Context, peerID string, content string) error
}
