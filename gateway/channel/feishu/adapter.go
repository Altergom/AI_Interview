package feishu

import (
	"context"
	"encoding/json"

	"gateway/inbound"
)

// Adapter 飞书渠道适配器骨架。
type Adapter struct{}

func New() *Adapter { return &Adapter{} }

func (a *Adapter) Name() string { return "feishu" }

func (a *Adapter) VerifySignature(_ context.Context, _ map[string]string, _ []byte) error {
	// TODO: 飞书签名验证（X-Lark-Signature HMAC-SHA256）
	return nil
}

func (a *Adapter) Parse(_ context.Context, body []byte) (*inbound.InboundEvent, error) {
	// TODO: 解析飞书事件回调格式
	return &inbound.InboundEvent{Channel: "feishu", Raw: body}, nil
}

func (a *Adapter) Ack(_ context.Context, body []byte) (any, error) {
	// 飞书 URL 验证时需原样返回 challenge
	var m map[string]any
	if err := json.Unmarshal(body, &m); err == nil {
		if challenge, ok := m["challenge"]; ok {
			return map[string]any{"challenge": challenge}, nil
		}
	}
	return map[string]any{}, nil
}

func (a *Adapter) Send(_ context.Context, _ string, _ string) error {
	// TODO: 调用飞书发送消息 API
	return nil
}
