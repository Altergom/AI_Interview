package weixin

import (
	"context"
	"log"
	"strconv"
	"sync"
	"time"

	"gateway/channel"
	"gateway/inbound"
)

// 重连退避，对照官方 RECONNECT_DELAYS 思路。
var reconnectDelays = []time.Duration{
	1 * time.Second,
	2 * time.Second,
	5 * time.Second,
	10 * time.Second,
	30 * time.Second,
}

// Connector 微信 iLink 渠道连接器，单账号一个实例。
// Start 内跑 getupdates 长轮询循环，把 message_type==USER 的消息归一为 InboundEvent。
type Connector struct {
	account *Account
	client  *Client

	mu        sync.RWMutex
	connected bool
	detail    string
}

// NewConnector 用已登录账号构造连接器。
func NewConnector(account *Account) *Connector {
	baseURL := account.BaseURL
	return &Connector{
		account: account,
		client:  NewClient(baseURL, account.Token),
		detail:  "initialized",
	}
}

func (c *Connector) Name() string { return "wechat" }

// Start 启动长轮询循环，阻塞直到 ctx 取消。内部自管重连退避。
func (c *Connector) Start(ctx context.Context, out chan<- *inbound.InboundEvent) error {
	var updatesBuf string
	attempt := 0
	for {
		if ctx.Err() != nil {
			c.setStatus(false, "stopped")
			return ctx.Err()
		}

		resp, err := c.client.GetUpdates(ctx, updatesBuf)
		if err != nil {
			if ctx.Err() != nil {
				c.setStatus(false, "stopped")
				return ctx.Err()
			}
			c.setStatus(false, "reconnecting")
			delay := backoff(attempt)
			attempt++
			log.Printf("[weixin] getupdates failed, retry in %s: %v", delay, err)
			select {
			case <-ctx.Done():
				c.setStatus(false, "stopped")
				return ctx.Err()
			case <-time.After(delay):
			}
			continue
		}

		attempt = 0
		c.setStatus(true, "connected")
		if resp.GetUpdatesBuf != "" {
			updatesBuf = resp.GetUpdatesBuf
		}

		for i := range resp.Msgs {
			ev := toInboundEvent(c.account.AccountID, &resp.Msgs[i])
			if ev == nil {
				continue
			}
			select {
			case <-ctx.Done():
				c.setStatus(false, "stopped")
				return ctx.Err()
			case out <- ev:
			}
		}
	}
}

// Send 出站发送文本。peerID 为对方 user id；opts.ContextToken 回传会话上下文。
func (c *Connector) Send(ctx context.Context, peerID, content string, opts channel.SendOpts) error {
	msg := &WeixinMessage{
		ToUserID:     peerID,
		MessageType:  MessageTypeBot,
		ContextToken: opts.ContextToken,
		ItemList: []MessageItem{
			{Type: MessageItemTypeText, TextItem: &TextItem{Text: content}},
		},
	}
	return c.client.SendMessage(ctx, msg)
}

func (c *Connector) Status() channel.ConnStatus {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return channel.ConnStatus{Connected: c.connected, Detail: c.detail}
}

func (c *Connector) setStatus(connected bool, detail string) {
	c.mu.Lock()
	c.connected = connected
	c.detail = detail
	c.mu.Unlock()
}

func backoff(attempt int) time.Duration {
	if attempt >= len(reconnectDelays) {
		return reconnectDelays[len(reconnectDelays)-1]
	}
	return reconnectDelays[attempt]
}

// toInboundEvent 把一条 iLink 消息归一为网关入站事件。
// 只处理用户发来的文本消息（message_type==USER 且含 text_item），其余返回 nil 跳过。
func toInboundEvent(accountID string, m *WeixinMessage) *inbound.InboundEvent {
	if m.MessageType != MessageTypeUser {
		return nil
	}
	text, ok := firstText(m.ItemList)
	if !ok {
		return nil
	}
	peerType := "user"
	if m.GroupID != "" {
		peerType = "group"
	}
	return &inbound.InboundEvent{
		Channel:   "wechat",
		AccountID: accountID,
		PeerID:    m.FromUserID,
		PeerType:  peerType,
		MessageID: strconv.FormatInt(m.MessageID, 10),
		Payload:   inbound.EventPayload{Type: "text", Content: text},
		ReceivedAt: time.Now(),
	}
}

func firstText(items []MessageItem) (string, bool) {
	for i := range items {
		if items[i].Type == MessageItemTypeText && items[i].TextItem != nil {
			return items[i].TextItem.Text, true
		}
	}
	return "", false
}
