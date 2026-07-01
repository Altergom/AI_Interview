package auth_test

import (
	"context"
	"encoding/json"
	"net/http"
	"testing"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/app/server"
	"github.com/cloudwego/hertz/pkg/common/ut"

	jwtauth "ai_interview/internal/auth"
	biz "ai_interview/internal/utils/respx"
	authmw "ai_interview/internal/middleware/auth"
)

const testSecret = "test-secret-32-bytes-long-enough!"

// newTestServer 创建测试用 Hertz server（不启动监听），注册 /protected 路由。
func newTestServer() *server.Hertz {
	h := server.New(server.WithHostPorts("127.0.0.1:0"))
	h.GET("/protected", authmw.HAuth(testSecret), func(ctx context.Context, c *app.RequestContext) {
		uid := authmw.GetUserID(c)
		guest := authmw.GetIsGuest(c)
		c.JSON(http.StatusOK, map[string]any{
			"user_id":  uid,
			"is_guest": guest,
		})
	})
	return h
}

func makeToken(t *testing.T, userID string, isGuest bool) string {
	t.Helper()
	cfg := jwtauth.TokenConfig{Secret: testSecret, Issuer: "test", ExpMinute: 60}
	tok, err := jwtauth.GenerateToken(cfg, userID, isGuest)
	if err != nil {
		t.Fatalf("GenerateToken: %v", err)
	}
	return tok
}

func TestHAuth_ValidToken(t *testing.T) {
	h := newTestServer()
	token := makeToken(t, "user-abc", false)

	w := ut.PerformRequest(h.Engine, "GET", "/protected", nil,
		ut.Header{Key: "Authorization", Value: "Bearer " + token})
	resp := w.Result()

	if resp.StatusCode() != http.StatusOK {
		t.Fatalf("expected 200, got %d body=%s", resp.StatusCode(), resp.Body())
	}
}

func TestHAuth_GuestToken(t *testing.T) {
	h := newTestServer()
	token := makeToken(t, "guest_abc12345", true)

	w := ut.PerformRequest(h.Engine, "GET", "/protected", nil,
		ut.Header{Key: "Authorization", Value: "Bearer " + token})
	resp := w.Result()

	if resp.StatusCode() != http.StatusOK {
		t.Fatalf("guest token should pass, got %d", resp.StatusCode())
	}
}

func TestHAuth_MissingToken(t *testing.T) {
	h := newTestServer()

	w := ut.PerformRequest(h.Engine, "GET", "/protected", nil)
	resp := w.Result()

	if resp.StatusCode() != http.StatusOK {
		t.Fatalf("expected 200, got %d", resp.StatusCode())
	}
	var body struct {
		Code int    `json:"code"`
		Msg  string `json:"msg"`
		Data any    `json:"data"`
	}
	if err := json.Unmarshal(resp.Body(), &body); err != nil {
		t.Fatalf("unmarshal body: %v, raw=%s", err, resp.Body())
	}
	if body.Code != int(biz.CodeUnauthorized) {
		t.Fatalf("expected code %d, got %d", int(biz.CodeUnauthorized), body.Code)
	}
	if body.Msg == "" {
		t.Fatalf("expected non-empty msg")
	}
}

func TestHAuth_InvalidToken(t *testing.T) {
	h := newTestServer()

	w := ut.PerformRequest(h.Engine, "GET", "/protected", nil,
		ut.Header{Key: "Authorization", Value: "Bearer this.is.invalid"})
	resp := w.Result()

	if resp.StatusCode() != http.StatusOK {
		t.Fatalf("expected 200, got %d", resp.StatusCode())
	}
	var body struct {
		Code int    `json:"code"`
		Msg  string `json:"msg"`
		Data any    `json:"data"`
	}
	if err := json.Unmarshal(resp.Body(), &body); err != nil {
		t.Fatalf("unmarshal body: %v, raw=%s", err, resp.Body())
	}
	if body.Code != int(biz.CodeUnauthorized) {
		t.Fatalf("expected code %d, got %d", int(biz.CodeUnauthorized), body.Code)
	}
	if body.Msg == "" {
		t.Fatalf("expected non-empty msg")
	}
}

func TestHAuth_WrongSecret(t *testing.T) {
	h := newTestServer()
	cfg := jwtauth.TokenConfig{Secret: "another-secret-32-bytes-padding!!", Issuer: "x", ExpMinute: 60}
	tok, _ := jwtauth.GenerateToken(cfg, "user-xyz", false)

	w := ut.PerformRequest(h.Engine, "GET", "/protected", nil,
		ut.Header{Key: "Authorization", Value: "Bearer " + tok})
	resp := w.Result()

	if resp.StatusCode() != http.StatusOK {
		t.Fatalf("expected 200 for wrong secret, got %d", resp.StatusCode())
	}
	var body struct {
		Code int    `json:"code"`
		Msg  string `json:"msg"`
		Data any    `json:"data"`
	}
	if err := json.Unmarshal(resp.Body(), &body); err != nil {
		t.Fatalf("unmarshal body: %v, raw=%s", err, resp.Body())
	}
	if body.Code != int(biz.CodeUnauthorized) {
		t.Fatalf("expected code %d, got %d", int(biz.CodeUnauthorized), body.Code)
	}
	if body.Msg == "" {
		t.Fatalf("expected non-empty msg")
	}
}
