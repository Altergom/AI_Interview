package handler

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/cloudwego/hertz/pkg/common/ut"

	"gateway/channel/weixin"
)

// loginManagerTo 构造一个 LoginManager，其内部 client 指向给定 httptest server。
func loginManagerTo(url string) *weixin.LoginManager {
	m := weixin.NewLoginManager(weixin.NewStore())
	m.SetClientFactory(func(string) *weixin.Client { return weixin.NewClient(url, "") })
	return m
}

// decodeResult 解析统一响应体。
func decodeResult(t *testing.T, body []byte) Result {
	t.Helper()
	var r Result
	if err := json.Unmarshal(body, &r); err != nil {
		t.Fatalf("响应非 JSON: %v body=%s", err, string(body))
	}
	return r
}

func TestWeixinLoginStart(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(map[string]string{
			"qrcode": "QR1", "qrcode_img_content": "https://q/1",
		})
	}))
	defer srv.Close()

	h := NewWeixinLoginHandler(loginManagerTo(srv.URL))
	c := ut.CreateUtRequestContext(http.MethodPost, "/v1/gateway/weixin/login", nil)
	h.Start(context.Background(), c)

	if code := c.Response.StatusCode(); code != http.StatusOK {
		t.Fatalf("status=%d, 期望 200, body=%s", code, c.Response.Body())
	}
	res := decodeResult(t, c.Response.Body())
	if !res.Success {
		t.Fatalf("success=false: %+v", res.Error)
	}
	data := res.Data.(map[string]any)
	if data["qrcode_url"] != "https://q/1" {
		t.Errorf("qrcode_url 错误: %v", data["qrcode_url"])
	}
	if data["session_key"] == "" || data["session_key"] == nil {
		t.Error("缺少 session_key")
	}
}

func TestWeixinLoginStartUpstreamError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "boom", http.StatusInternalServerError)
	}))
	defer srv.Close()

	h := NewWeixinLoginHandler(loginManagerTo(srv.URL))
	c := ut.CreateUtRequestContext(http.MethodPost, "/v1/gateway/weixin/login", nil)
	h.Start(context.Background(), c)

	if code := c.Response.StatusCode(); code != http.StatusBadGateway {
		t.Fatalf("上游失败应返回 502, got %d", code)
	}
	res := decodeResult(t, c.Response.Body())
	if res.Success || res.Error == nil {
		t.Fatalf("应为失败响应: %+v", res)
	}
}

func TestWeixinLoginStatusMissingKey(t *testing.T) {
	h := NewWeixinLoginHandler(loginManagerTo("http://unused"))
	c := ut.CreateUtRequestContext(http.MethodGet, "/v1/gateway/weixin/login/status", nil)
	h.Status(context.Background(), c)

	if code := c.Response.StatusCode(); code != http.StatusBadRequest {
		t.Fatalf("缺 session_key 应返回 400, got %d", code)
	}
}

func TestWeixinLoginStatusNotFound(t *testing.T) {
	h := NewWeixinLoginHandler(loginManagerTo("http://unused"))
	c := ut.CreateUtRequestContext(http.MethodGet, "/v1/gateway/weixin/login/status?session_key=ghost", nil)
	h.Status(context.Background(), c)

	if code := c.Response.StatusCode(); code != http.StatusNotFound {
		t.Fatalf("未知 session 应返回 404, got %d", code)
	}
}

func TestWeixinLoginStatusFound(t *testing.T) {
	// 仅返回 wait，避免后台轮询很快结束；这里只验证 status 接口能读到会话。
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case strings.Contains(r.URL.Path, "get_bot_qrcode"):
			json.NewEncoder(w).Encode(map[string]string{"qrcode": "QR1", "qrcode_img_content": "https://q/1"})
		default:
			json.NewEncoder(w).Encode(map[string]string{"status": "wait"})
		}
	}))
	defer srv.Close()

	h := NewWeixinLoginHandler(loginManagerTo(srv.URL))
	startC := ut.CreateUtRequestContext(http.MethodPost, "/v1/gateway/weixin/login", nil)
	h.Start(context.Background(), startC)
	sessionKey := decodeResult(t, startC.Response.Body()).Data.(map[string]any)["session_key"].(string)

	statusC := ut.CreateUtRequestContext(http.MethodGet, "/v1/gateway/weixin/login/status?session_key="+sessionKey, nil)
	h.Status(context.Background(), statusC)

	if code := statusC.Response.StatusCode(); code != http.StatusOK {
		t.Fatalf("已知 session 应返回 200, got %d", code)
	}
	res := decodeResult(t, statusC.Response.Body())
	if !res.Success || res.Data.(map[string]any)["status"] != "wait" {
		t.Fatalf("status 响应错误: %+v", res.Data)
	}
}
