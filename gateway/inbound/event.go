package inbound

import "time"

// InboundEvent 统一入站事件，所有渠道消息转换为此结构后进入网关处理流程。
// 平台特有字段只放 Raw，不进主协议。
type InboundEvent struct {
	Channel    string       // wechat / feishu / qqbot
	AccountID  string       // 我们的平台账号 ID（区分多账号）
	PeerID     string       // 用户在平台的唯一 ID
	PeerType   string       // user / group
	MessageID  string       // 平台消息 ID（去重用）
	Payload    EventPayload
	Raw        []byte // 原始平台数据
	ReceivedAt time.Time
}

// EventPayload 统一消息内容，支持 text / image / audio。
type EventPayload struct {
	Type    string // text / image / audio
	Content string // text 时为文本内容，其他类型为资源 URL 或 base64
}
