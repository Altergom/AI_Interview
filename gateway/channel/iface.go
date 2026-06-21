package channel

import (
	"context"

	"gateway/inbound"
)

// ChannelConnector 渠道连接器。
// 每个渠道账号对应一个常驻连接（长轮询 / WebSocket / webhook 由各实现自行管理），
// 负责维持入站连接、归一消息为 InboundEvent，并提供出站发送能力。
//
// 三个目标渠道（微信 iLink 长轮询、飞书 WebSocket、QQ WebSocket/webhook）的接入模式
// 都收敛到「主动维持一条长连接 + 自管重连 + 取消即停」的模型，故抽象为此接口，
// 取代早期按 webhook 被动推送设计的 ChannelAdapter。
type ChannelConnector interface {
	// Name 返回渠道标识，如 "wechat" / "feishu" / "qqbot"。
	Name() string

	// Start 启动该渠道的入站连接，把收到的消息归一为 InboundEvent 投入 out。
	// 阻塞直到 ctx 取消；内部自行管理重连、心跳、退避。
	Start(ctx context.Context, out chan<- *inbound.InboundEvent) error

	// Send 向平台用户发送消息。
	Send(ctx context.Context, peerID, content string, opts SendOpts) error

	// Status 返回当前连接状态，供健康检查使用。
	Status() ConnStatus
}

// SendOpts 出站发送的附加选项。
type SendOpts struct {
	// ContextToken 会话上下文令牌。微信回复时需回传；其他渠道可空。
	ContextToken string
	// AccountID 多账号场景下指定发送账号；单账号可空。
	AccountID string
}

// ConnStatus 渠道连接状态。
type ConnStatus struct {
	Connected bool   // 是否已建立连接
	Detail    string // 状态描述（如 "reconnecting"、"not implemented"）
}
