package outbound

import (
	"context"
	"fmt"

	"gateway/channel"
)

// Dispatcher 出站分发器，按 channel 把网关动作回写到对应平台。
type Dispatcher struct {
	connectors map[string]channel.ChannelConnector
}

func NewDispatcher(connectors ...channel.ChannelConnector) *Dispatcher {
	m := make(map[string]channel.ChannelConnector, len(connectors))
	for _, c := range connectors {
		m[c.Name()] = c
	}
	return &Dispatcher{connectors: m}
}

// Send 向指定渠道的用户发送文本消息。
func (d *Dispatcher) Send(ctx context.Context, ch, peerID, content string, opts channel.SendOpts) error {
	conn, ok := d.connectors[ch]
	if !ok {
		return fmt.Errorf("unknown channel: %s", ch)
	}
	return conn.Send(ctx, peerID, content, opts)
}
