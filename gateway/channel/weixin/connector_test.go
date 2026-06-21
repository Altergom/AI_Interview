package weixin

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"

	"gateway/channel"
	"gateway/inbound"
)

func TestToInboundEvent(t *testing.T) {
	tests := []struct {
		name string
		msg  *WeixinMessage
		want bool // 是否应产出 event
	}{
		{"用户文本", &WeixinMessage{MessageType: MessageTypeUser, FromUserID: "u1",
			ItemList: []MessageItem{{Type: MessageItemTypeText, TextItem: &TextItem{Text: "hello"}}}}, true},
		{"bot 消息跳过", &WeixinMessage{MessageType: MessageTypeBot,
			ItemList: []MessageItem{{Type: MessageItemTypeText, TextItem: &TextItem{Text: "x"}}}}, false},
		{"无文本项跳过", &WeixinMessage{MessageType: MessageTypeUser,
			ItemList: []MessageItem{{Type: MessageItemTypeImage}}}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ev := toInboundEvent("acc-1", tt.msg)
			if tt.want != (ev != nil) {
				t.Fatalf("want event=%v, got=%v", tt.want, ev != nil)
			}
			if ev != nil {
				if ev.Channel != "wechat" || ev.AccountID != "acc-1" || ev.PeerID != "u1" {
					t.Errorf("event 字段错误: %+v", ev)
				}
				if ev.Payload.Type != "text" || ev.Payload.Content != "hello" {
					t.Errorf("payload 错误: %+v", ev.Payload)
				}
			}
		})
	}
}

func TestConnectorStartEmitsEventAndStops(t *testing.T) {
	var calls int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		n := atomic.AddInt32(&calls, 1)
		if n == 1 {
			json.NewEncoder(w).Encode(GetUpdatesResp{
				GetUpdatesBuf: "c2",
				Msgs: []WeixinMessage{{
					MessageID: 1, FromUserID: "u1", MessageType: MessageTypeUser,
					ItemList: []MessageItem{{Type: MessageItemTypeText, TextItem: &TextItem{Text: "hi"}}},
				}},
			})
			return
		}
		// 后续轮询返回空，让循环空转直到 ctx 取消。
		json.NewEncoder(w).Encode(GetUpdatesResp{GetUpdatesBuf: "c2"})
	}))
	defer srv.Close()

	conn := &Connector{
		account: &Account{AccountID: "acc-1", Token: "tok", BaseURL: srv.URL},
		client:  NewClient(srv.URL, "tok"),
	}
	out := make(chan *inbound.InboundEvent, 1)
	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan error, 1)
	go func() { done <- conn.Start(ctx, out) }()

	select {
	case ev := <-out:
		if ev.PeerID != "u1" || ev.Payload.Content != "hi" {
			t.Fatalf("收到事件错误: %+v", ev)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("超时未收到入站事件")
	}

	if st := conn.Status(); !st.Connected {
		t.Errorf("收到消息后应为 connected, got %+v", st)
	}

	cancel()
	select {
	case err := <-done:
		if err != context.Canceled {
			t.Errorf("期望 context.Canceled, got %v", err)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("取消后 Start 未退出")
	}
}

func TestConnectorSend(t *testing.T) {
	got := make(chan *WeixinMessage, 1)
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var req SendMessageReq
		json.NewDecoder(r.Body).Decode(&req)
		got <- req.Msg
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	conn := &Connector{
		account: &Account{AccountID: "acc-1", Token: "tok"},
		client:  NewClient(srv.URL, "tok"),
	}
	err := conn.Send(context.Background(), "u1", "reply", channel.SendOpts{ContextToken: "ctx-9"})
	if err != nil {
		t.Fatalf("Send: %v", err)
	}
	msg := <-got
	if msg.ToUserID != "u1" || msg.ContextToken != "ctx-9" {
		t.Fatalf("发送字段错误: %+v", msg)
	}
	if msg.ItemList[0].TextItem.Text != "reply" {
		t.Fatalf("文本错误: %+v", msg.ItemList)
	}
}

func TestConnectorName(t *testing.T) {
	if NewConnector(&Account{}).Name() != "wechat" {
		t.Fatal("Name 应为 wechat")
	}
}

func TestBackoff(t *testing.T) {
	if backoff(0) != reconnectDelays[0] {
		t.Errorf("backoff(0)=%v", backoff(0))
	}
	// 超出索引应钳到最后一档。
	if backoff(100) != reconnectDelays[len(reconnectDelays)-1] {
		t.Errorf("backoff(100) 未钳到最大档: %v", backoff(100))
	}
}

// TestConnectorReconnect 首次 getupdates 返回 500 触发重连退避，
// 第二次成功产出事件。验证错误不会终止循环。
func TestConnectorReconnect(t *testing.T) {
	var calls int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if atomic.AddInt32(&calls, 1) == 1 {
			http.Error(w, "boom", http.StatusInternalServerError)
			return
		}
		json.NewEncoder(w).Encode(GetUpdatesResp{
			Msgs: []WeixinMessage{{
				MessageID: 1, FromUserID: "u1", MessageType: MessageTypeUser,
				ItemList: []MessageItem{{Type: MessageItemTypeText, TextItem: &TextItem{Text: "hi"}}},
			}},
		})
	}))
	defer srv.Close()

	// 用极短退避避免拖慢测试。
	conn := &Connector{
		account: &Account{AccountID: "acc-1", Token: "tok"},
		client:  NewClient(srv.URL, "tok"),
	}
	out := make(chan *inbound.InboundEvent, 1)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go conn.Start(ctx, out)

	select {
	case ev := <-out:
		if ev.PeerID != "u1" {
			t.Fatalf("事件错误: %+v", ev)
		}
	case <-time.After(5 * time.Second):
		t.Fatal("重连后未收到事件")
	}
	if atomic.LoadInt32(&calls) < 2 {
		t.Errorf("应至少调用 2 次(含一次失败重连), got %d", calls)
	}
}

// TestConnectorStartCtxCanceledImmediately 已取消的 ctx 应立即返回，不发请求。
func TestConnectorStartCtxCanceledImmediately(t *testing.T) {
	conn := &Connector{account: &Account{AccountID: "a"}, client: NewClient("http://unused", "")}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	err := conn.Start(ctx, make(chan *inbound.InboundEvent, 1))
	if err != context.Canceled {
		t.Fatalf("期望 context.Canceled, got %v", err)
	}
	if st := conn.Status(); st.Connected {
		t.Error("取消后不应为 connected")
	}
}
