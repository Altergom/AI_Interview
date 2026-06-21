package qqbot

import (
	"context"
	"fmt"

	"gateway/channel"
	"gateway/inbound"
)

// Connector QQ Bot 渠道连接器骨架。
// QQ 官方支持 WebSocket（默认）/ Webhook 双 transport（tencent-connect/openclaw-qqbot），
// 此处为占位实现，待后续补全。
type Connector struct{}

func New() *Connector { return &Connector{} }

func (c *Connector) Name() string { return "qqbot" }

// Start 尚未实现。QQ 默认建立 WebSocket 长连接接收事件。
func (c *Connector) Start(_ context.Context, _ chan<- *inbound.InboundEvent) error {
	return fmt.Errorf("qqbot: connector not implemented")
}

func (c *Connector) Send(_ context.Context, _ string, _ string, _ channel.SendOpts) error {
	return fmt.Errorf("qqbot: send not implemented")
}

func (c *Connector) Status() channel.ConnStatus {
	return channel.ConnStatus{Connected: false, Detail: "not implemented"}
}
