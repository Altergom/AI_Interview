package outbound

import (
	"context"
	"fmt"

	"gateway/channel"
)

// Dispatcher 出站分发器，按 channel 把网关动作回写到对应平台。
type Dispatcher struct {
	adapters map[string]channel.ChannelAdapter
}

func NewDispatcher(adapters ...channel.ChannelAdapter) *Dispatcher {
	m := make(map[string]channel.ChannelAdapter, len(adapters))
	for _, a := range adapters {
		m[a.Name()] = a
	}
	return &Dispatcher{adapters: m}
}

// Send 向指定渠道的用户发送文本消息。
func (d *Dispatcher) Send(ctx context.Context, ch, peerID, content string) error {
	adapter, ok := d.adapters[ch]
	if !ok {
		return fmt.Errorf("unknown channel: %s", ch)
	}
	return adapter.Send(ctx, peerID, content)
}
