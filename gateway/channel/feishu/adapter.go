package feishu

import (
	"context"
	"fmt"

	"gateway/channel"
	"gateway/inbound"
)

// Connector 飞书渠道连接器骨架。
// 飞书官方走 WebSocket 长连接（larksuite/openclaw-lark），此处为占位实现，
// 待后续按 WSClient 模型补全。
type Connector struct{}

func New() *Connector { return &Connector{} }

func (c *Connector) Name() string { return "feishu" }

// Start 尚未实现。飞书需建立 WebSocket 长连接订阅 im.message.receive_v1 等事件。
func (c *Connector) Start(_ context.Context, _ chan<- *inbound.InboundEvent) error {
	return fmt.Errorf("feishu: connector not implemented")
}

func (c *Connector) Send(_ context.Context, _ string, _ string, _ channel.SendOpts) error {
	return fmt.Errorf("feishu: send not implemented")
}

func (c *Connector) Status() channel.ConnStatus {
	return channel.ConnStatus{Connected: false, Detail: "not implemented"}
}
