package weixin

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"
	"time"
)

func TestStoreSaveGetTokens(t *testing.T) {
	s := NewStore()
	s.Save(&Account{AccountID: "a1", Token: "t1"})
	s.Save(&Account{AccountID: "a2", Token: "t2"})

	if a, ok := s.Get("a1"); !ok || a.Token != "t1" {
		t.Fatalf("Get a1 错误: %+v ok=%v", a, ok)
	}
	if _, ok := s.Get("missing"); ok {
		t.Fatal("不存在账号应返回 ok=false")
	}
	if len(s.List()) != 2 {
		t.Fatalf("List 应有 2 个, got %d", len(s.List()))
	}
	if len(s.Tokens()) != 2 {
		t.Fatalf("Tokens 应有 2 个, got %d", len(s.Tokens()))
	}
}

// loginTestServer 模拟 get_bot_qrcode + get_qrcode_status，status 按调用次数推进。
func loginTestServer(t *testing.T, statuses []StatusResp) *httptest.Server {
	t.Helper()
	var statusCalls int32
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case strings.Contains(r.URL.Path, epGetBotQRCode):
			json.NewEncoder(w).Encode(QRCodeResp{QRCode: "QR1", QRCodeImgContent: "https://q/1"})
		case strings.Contains(r.URL.Path, epGetQRCodeStatus):
			i := int(atomic.AddInt32(&statusCalls, 1)) - 1
			if i >= len(statuses) {
				i = len(statuses) - 1
			}
			json.NewEncoder(w).Encode(statuses[i])
		default:
			t.Errorf("意外端点: %s", r.URL.Path)
		}
	}))
}

func TestLoginManagerConfirmed(t *testing.T) {
	srv := loginTestServer(t, []StatusResp{
		{Status: QRStatusWait},
		{Status: QRStatusScaned},
		{Status: QRStatusConfirmed, BotToken: "tok", ILinkBotID: "bot-1", ILinkUserID: "user-1"},
	})
	defer srv.Close()

	store := NewStore()
	m := NewLoginManager(store)
	m.SetClientFactory(func(string) *Client { return NewClient(srv.URL, "") })

	sess, err := m.Start(context.Background(), "key-1")
	if err != nil {
		t.Fatalf("Start: %v", err)
	}
	if sess.QRCodeURL != "https://q/1" {
		t.Fatalf("二维码 URL 错误: %s", sess.QRCodeURL)
	}

	// 轮询间隔 1s，等足够时间走完 3 次状态。
	deadline := time.Now().Add(6 * time.Second)
	for time.Now().Before(deadline) {
		if st, _ := m.Status("key-1"); st != nil && st.Status == QRStatusConfirmed {
			break
		}
		time.Sleep(100 * time.Millisecond)
	}

	st, _ := m.Status("key-1")
	if st.Status != QRStatusConfirmed || st.AccountID != "bot-1" {
		t.Fatalf("登录未确认: %+v", st)
	}
	if a, ok := store.Get("bot-1"); !ok || a.Token != "tok" || a.UserID != "user-1" {
		t.Fatalf("token 未落库: %+v ok=%v", a, ok)
	}
}

// TestLoginManagerStartReuse 同 sessionKey 在 fresh 期内重复 Start 应复用，不再拉新码。
func TestLoginManagerStartReuse(t *testing.T) {
	var qrCalls int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case strings.Contains(r.URL.Path, epGetBotQRCode):
			atomic.AddInt32(&qrCalls, 1)
			json.NewEncoder(w).Encode(QRCodeResp{QRCode: "QR1", QRCodeImgContent: "https://q/1"})
		default:
			json.NewEncoder(w).Encode(StatusResp{Status: QRStatusWait})
		}
	}))
	defer srv.Close()

	m := NewLoginManager(NewStore())
	m.SetClientFactory(func(string) *Client { return NewClient(srv.URL, "") })

	s1, err := m.Start(context.Background(), "k")
	if err != nil {
		t.Fatalf("Start1: %v", err)
	}
	s2, err := m.Start(context.Background(), "k")
	if err != nil {
		t.Fatalf("Start2: %v", err)
	}
	if s1.QRCode != s2.QRCode {
		t.Error("复用应返回同一二维码")
	}
	if n := atomic.LoadInt32(&qrCalls); n != 1 {
		t.Errorf("fresh 期内应只拉一次码, got %d", n)
	}
}

func TestLoginManagerExpired(t *testing.T) {
	srv := loginTestServer(t, []StatusResp{{Status: QRStatusExpired}})
	defer srv.Close()

	m := NewLoginManager(NewStore())
	m.SetClientFactory(func(string) *Client { return NewClient(srv.URL, "") })

	if _, err := m.Start(context.Background(), "key-1"); err != nil {
		t.Fatalf("Start: %v", err)
	}

	deadline := time.Now().Add(3 * time.Second)
	for time.Now().Before(deadline) {
		if st, _ := m.Status("key-1"); st != nil && st.Err != "" {
			break
		}
		time.Sleep(100 * time.Millisecond)
	}
	st, _ := m.Status("key-1")
	if st.Err == "" {
		t.Fatalf("过期应记录 Err, got %+v", st)
	}
}

func waitErr(t *testing.T, m *LoginManager, key string) *LoginSession {
	t.Helper()
	deadline := time.Now().Add(3 * time.Second)
	for time.Now().Before(deadline) {
		if st, _ := m.Status(key); st != nil && st.Err != "" {
			return st
		}
		time.Sleep(100 * time.Millisecond)
	}
	st, _ := m.Status(key)
	return st
}

func TestLoginManagerBindedRedirect(t *testing.T) {
	srv := loginTestServer(t, []StatusResp{{Status: QRStatusBindedRedirect}})
	defer srv.Close()
	m := NewLoginManager(NewStore())
	m.SetClientFactory(func(string) *Client { return NewClient(srv.URL, "") })
	if _, err := m.Start(context.Background(), "k"); err != nil {
		t.Fatalf("Start: %v", err)
	}
	if st := waitErr(t, m, "k"); st.Err == "" {
		t.Fatalf("binded_redirect 应记录 Err, got %+v", st)
	}
}

func TestLoginManagerNeedVerifyCode(t *testing.T) {
	srv := loginTestServer(t, []StatusResp{{Status: QRStatusNeedVerifyCode}})
	defer srv.Close()
	m := NewLoginManager(NewStore())
	m.SetClientFactory(func(string) *Client { return NewClient(srv.URL, "") })
	if _, err := m.Start(context.Background(), "k"); err != nil {
		t.Fatalf("Start: %v", err)
	}
	if st := waitErr(t, m, "k"); st.Err == "" {
		t.Fatalf("need_verifycode 应记录 Err（本期不支持）, got %+v", st)
	}
}

// TestLoginManagerScanedButRedirect 先返回切主机指令，再在新主机确认登录。
func TestLoginManagerScanedButRedirect(t *testing.T) {
	// 第二台 server 作为重定向目标，确认登录。
	redirectSrv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(StatusResp{
			Status: QRStatusConfirmed, BotToken: "tok2", ILinkBotID: "bot-2", ILinkUserID: "u2",
		})
	}))
	defer redirectSrv.Close()
	redirectHost := strings.TrimPrefix(redirectSrv.URL, "http://")

	var statusCalls int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case strings.Contains(r.URL.Path, epGetBotQRCode):
			json.NewEncoder(w).Encode(QRCodeResp{QRCode: "QR1", QRCodeImgContent: "https://q/1"})
		case strings.Contains(r.URL.Path, epGetQRCodeStatus):
			atomic.AddInt32(&statusCalls, 1)
			json.NewEncoder(w).Encode(StatusResp{Status: QRStatusScanedButRedirect, RedirectHost: redirectHost})
		}
	}))
	defer srv.Close()

	store := NewStore()
	m := NewLoginManager(store)
	// 工厂按 baseURL 路由：重定向后会用 http://<redirectHost> 构造 client。
	m.SetClientFactory(func(baseURL string) *Client {
		if strings.Contains(baseURL, redirectHost) {
			return NewClient(redirectSrv.URL, "")
		}
		return NewClient(srv.URL, "")
	})
	if _, err := m.Start(context.Background(), "k"); err != nil {
		t.Fatalf("Start: %v", err)
	}

	deadline := time.Now().Add(6 * time.Second)
	for time.Now().Before(deadline) {
		if st, _ := m.Status("k"); st != nil && st.Status == QRStatusConfirmed {
			break
		}
		time.Sleep(100 * time.Millisecond)
	}
	st, _ := m.Status("k")
	if st.Status != QRStatusConfirmed || st.AccountID != "bot-2" {
		t.Fatalf("重定向后应在新主机确认: %+v", st)
	}
	if _, ok := store.Get("bot-2"); !ok {
		t.Fatal("重定向后 token 未落库")
	}
}
