package qqbot

import (
	"context"

	"gateway/inbound"
)

// Adapter QQ Bot 渠道适配器骨架。
type Adapter struct{}

func New() *Adapter { return &Adapter{} }

func (a *Adapter) Name() string { return "qqbot" }

func (a *Adapter) VerifySignature(_ context.Context, _ map[string]string, _ []byte) error {
	// TODO: QQ Bot 签名验证
	return nil
}

func (a *Adapter) Parse(_ context.Context, body []byte) (*inbound.InboundEvent, error) {
	// TODO: 解析 QQ Bot 消息格式
	return &inbound.InboundEvent{Channel: "qqbot", Raw: body}, nil
}

func (a *Adapter) Ack(_ context.Context, _ []byte) (any, error) {
	return map[string]any{"ret": 0}, nil
}

func (a *Adapter) Send(_ context.Context, _ string, _ string) error {
	// TODO: 调用 QQ Bot 发送消息 API
	return nil
}
