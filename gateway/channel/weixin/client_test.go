package weixin

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"
	"time"
)

// newTestClient 指向 httptest server 的客户端。
func newTestClient(t *testing.T, srv *httptest.Server, token string) *Client {
	t.Helper()
	return NewClient(srv.URL, token)
}

func TestRandomWechatUin(t *testing.T) {
	uin, err := randomWechatUin()
	if err != nil {
		t.Fatalf("randomWechatUin: %v", err)
	}
	raw, err := base64.StdEncoding.DecodeString(uin)
	if err != nil {
		t.Fatalf("uin 不是合法 base64: %v", err)
	}
	if _, err := strconv.ParseUint(string(raw), 10, 32); err != nil {
		t.Fatalf("解码后不是 uint32 十进制串: %q", string(raw))
	}
}

func TestGetBotQRCode(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !strings.Contains(r.URL.Path, epGetBotQRCode) {
			t.Errorf("意外端点: %s", r.URL.Path)
		}
		if r.URL.Query().Get("bot_type") != defaultBotType {
			t.Errorf("bot_type=%s, 期望 %s", r.URL.Query().Get("bot_type"), defaultBotType)
		}
		if r.Header.Get("iLink-App-Id") != ilinkAppID {
			t.Errorf("缺少 iLink-App-Id 头")
		}
		body, _ := io.ReadAll(r.Body)
		var req qrCodeReq
		if err := json.Unmarshal(body, &req); err != nil {
			t.Errorf("请求体非法: %v", err)
		}
		json.NewEncoder(w).Encode(QRCodeResp{QRCode: "QR123", QRCodeImgContent: "https://weixin.qq.com/q/abc"})
	}))
	defer srv.Close()

	cli := newTestClient(t, srv, "")
	resp, err := cli.GetBotQRCode(context.Background(), nil)
	if err != nil {
		t.Fatalf("GetBotQRCode: %v", err)
	}
	if resp.QRCode != "QR123" || resp.QRCodeImgContent != "https://weixin.qq.com/q/abc" {
		t.Fatalf("响应解析错误: %+v", resp)
	}
}

func TestGetQRCodeStatusConfirmed(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("get_qrcode_status 应为 GET, 实际 %s", r.Method)
		}
		if r.URL.Query().Get("qrcode") != "QR123" {
			t.Errorf("qrcode 透传错误: %s", r.URL.Query().Get("qrcode"))
		}
		json.NewEncoder(w).Encode(StatusResp{
			Status: QRStatusConfirmed, BotToken: "tok", ILinkBotID: "bot-1", ILinkUserID: "user-1",
		})
	}))
	defer srv.Close()

	cli := newTestClient(t, srv, "")
	resp, err := cli.GetQRCodeStatus(context.Background(), "QR123")
	if err != nil {
		t.Fatalf("GetQRCodeStatus: %v", err)
	}
	if resp.Status != QRStatusConfirmed || resp.BotToken != "tok" || resp.ILinkBotID != "bot-1" {
		t.Fatalf("状态解析错误: %+v", resp)
	}
}

func TestGetUpdatesAuthHeader(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if got := r.Header.Get("Authorization"); got != "Bearer mytoken" {
			t.Errorf("Authorization=%q, 期望 Bearer mytoken", got)
		}
		if r.Header.Get("AuthorizationType") != "ilink_bot_token" {
			t.Errorf("缺少 AuthorizationType 头")
		}
		if r.Header.Get("X-WECHAT-UIN") == "" {
			t.Errorf("缺少 X-WECHAT-UIN 头")
		}
		json.NewEncoder(w).Encode(GetUpdatesResp{
			Ret:           0,
			GetUpdatesBuf: "cursor-2",
			Msgs: []WeixinMessage{{
				MessageID:   42,
				FromUserID:  "user-1",
				MessageType: MessageTypeUser,
				ItemList:    []MessageItem{{Type: MessageItemTypeText, TextItem: &TextItem{Text: "hi"}}},
			}},
		})
	}))
	defer srv.Close()

	cli := newTestClient(t, srv, "mytoken")
	resp, err := cli.GetUpdates(context.Background(), "cursor-1")
	if err != nil {
		t.Fatalf("GetUpdates: %v", err)
	}
	if resp.GetUpdatesBuf != "cursor-2" || len(resp.Msgs) != 1 {
		t.Fatalf("响应解析错误: %+v", resp)
	}
}

func TestSendMessage(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		var req SendMessageReq
		if err := json.Unmarshal(body, &req); err != nil {
			t.Errorf("请求体非法: %v", err)
		}
		if req.Msg == nil || req.Msg.ToUserID != "user-1" {
			t.Errorf("msg 字段错误: %+v", req.Msg)
		}
		if len(req.Msg.ItemList) != 1 || req.Msg.ItemList[0].TextItem.Text != "reply" {
			t.Errorf("item_list 错误: %+v", req.Msg.ItemList)
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	cli := newTestClient(t, srv, "tok")
	err := cli.SendMessage(context.Background(), &WeixinMessage{
		ToUserID:    "user-1",
		MessageType: MessageTypeBot,
		ItemList:    []MessageItem{{Type: MessageItemTypeText, TextItem: &TextItem{Text: "reply"}}},
	})
	if err != nil {
		t.Fatalf("SendMessage: %v", err)
	}
}

func TestDoPostNon2xx(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "boom", http.StatusInternalServerError)
	}))
	defer srv.Close()

	cli := newTestClient(t, srv, "tok")
	_, err := cli.GetBotQRCode(context.Background(), nil)
	if err == nil {
		t.Fatal("期望 500 返回错误，实际为 nil")
	}
}

// TestIsPollTimeout 验证长轮询超时判定：内层 deadline 超时而外层存活才算。
func TestIsPollTimeout(t *testing.T) {
	expired, cancel := context.WithDeadline(context.Background(), time.Now().Add(-time.Second))
	defer cancel()
	if !isPollTimeout(context.Background(), expired) {
		t.Error("内层超时+外层存活应判为 poll timeout")
	}
	// 外层已取消：不算 poll timeout（应上抛取消）。
	outerCanceled, oc := context.WithCancel(context.Background())
	oc()
	if isPollTimeout(outerCanceled, expired) {
		t.Error("外层取消时不应判为 poll timeout")
	}
	// 内层未超时：不算。
	if isPollTimeout(context.Background(), context.Background()) {
		t.Error("内层未超时不应判为 poll timeout")
	}
}

// TestGetUpdatesHTTPErrorPropagates HTTP 5xx 必须上抛（触发重连），不能被当成空轮询吞掉。
func TestGetUpdatesHTTPErrorPropagates(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "boom", http.StatusInternalServerError)
	}))
	defer srv.Close()

	cli := newTestClient(t, srv, "tok")
	_, err := cli.GetUpdates(context.Background(), "cursor")
	if err == nil {
		t.Fatal("HTTP 500 应返回错误以触发重连，实际被吞掉")
	}
	if errors.Is(err, errPollTimeout) {
		t.Fatal("HTTP 500 不应被判为 poll timeout")
	}
}

// TestGetUpdatesPollTimeoutReturnsEmpty 客户端长轮询超时返回空响应并保留游标。
func TestGetUpdatesPollTimeoutReturnsEmpty(t *testing.T) {
	block := make(chan struct{})
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		<-block // 一直 hang，强制客户端超时
	}))
	defer srv.Close()
	defer close(block)

	// 用很短的内层 timeout 模拟长轮询超时。
	cli := NewClient(srv.URL, "tok")
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	resp, err := cli.getUpdatesWithTimeout(ctx, "cursor-1", 100*time.Millisecond)
	if err != nil {
		t.Fatalf("长轮询超时不应返回错误: %v", err)
	}
	if resp.GetUpdatesBuf != "cursor-1" {
		t.Fatalf("超时应保留游标, got %q", resp.GetUpdatesBuf)
	}
}

// TestGetQRCodeStatusPollTimeout 长轮询超时应返回 status=wait 继续等待。
func TestGetQRCodeStatusPollTimeout(t *testing.T) {
	block := make(chan struct{})
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		<-block
	}))
	defer srv.Close()
	defer close(block)

	// doGet 内层 timeout 100ms，外层 ctx 存活 → errPollTimeout。
	cli := NewClient(srv.URL, "")
	_, err := cli.doGet(context.Background(), epGetQRCodeStatus+"?qrcode=x", 100*time.Millisecond)
	if !errors.Is(err, errPollTimeout) {
		t.Fatalf("doGet 超时应返回 errPollTimeout, got %v", err)
	}
}

func TestSetToken(t *testing.T) {
	cli := NewClient("", "")
	cli.SetToken("abc")
	h := http.Header{}
	if err := cli.postHeaders(h); err != nil {
		t.Fatalf("postHeaders: %v", err)
	}
	if h.Get("Authorization") != "Bearer abc" {
		t.Fatalf("SetToken 后应带 Bearer 头, got %q", h.Get("Authorization"))
	}
}


