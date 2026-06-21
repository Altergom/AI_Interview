package weixin

import (
	"context"
	"fmt"
	"sync"
	"time"
)

// 登录会话默认参数，对照 openclaw-weixin login-qr.ts。
const (
	loginTTL      = 5 * time.Minute  // 二维码有效期
	loginDeadline = 8 * time.Minute  // 整个等待登录的最长时长
	pollInterval  = 1 * time.Second  // 两次轮询间隔
)

// LoginSession 一次扫码登录的状态。
type LoginSession struct {
	QRCode      string // get_qrcode_status 用的标识
	QRCodeURL   string // 二维码内容 URL，透传给前端渲染
	Status      string // 当前状态（wait/scaned/confirmed/...）
	AccountID   string // confirmed 后填充
	UserID      string // confirmed 后填充
	Err         string // 失败原因
	startedAt   time.Time
	baseURL     string // 当前有效的轮询主机，IDC 重定向时更新
}

func (l *LoginSession) fresh() bool { return time.Since(l.startedAt) < loginTTL }

// LoginManager 管理进行中的扫码登录会话，并在后台代为轮询状态。
type LoginManager struct {
	mu       sync.Mutex
	sessions map[string]*LoginSession // sessionKey → session
	store    *Store
	newClient func(baseURL string) *Client
}

// NewLoginManager 创建登录管理器。store 用于登录成功后落 token。
func NewLoginManager(store *Store) *LoginManager {
	return &LoginManager{
		sessions:  make(map[string]*LoginSession),
		store:     store,
		newClient: func(baseURL string) *Client { return NewClient(baseURL, "") },
	}
}

// SetClientFactory 替换内部 Client 工厂，仅供测试把请求指向 mock 后端。
func (m *LoginManager) SetClientFactory(f func(baseURL string) *Client) {
	m.newClient = f
}

// Start 发起一次登录：拉二维码，返回 sessionKey 和二维码 URL。
// 内部起后台 goroutine 轮询状态直到 confirmed/expired/超时。
func (m *LoginManager) Start(ctx context.Context, sessionKey string) (*LoginSession, error) {
	m.mu.Lock()
	if existing, ok := m.sessions[sessionKey]; ok && existing.fresh() && existing.QRCodeURL != "" {
		m.mu.Unlock()
		return existing, nil
	}
	m.mu.Unlock()

	cli := m.newClient("")
	qr, err := cli.GetBotQRCode(ctx, m.store.Tokens())
	if err != nil {
		return nil, fmt.Errorf("weixin: start login: %w", err)
	}
	sess := &LoginSession{
		QRCode:    qr.QRCode,
		QRCodeURL: qr.QRCodeImgContent,
		Status:    QRStatusWait,
		startedAt: time.Now(),
		baseURL:   defaultBaseURL,
	}
	m.mu.Lock()
	m.sessions[sessionKey] = sess
	m.mu.Unlock()

	go m.poll(context.Background(), sess)
	return sess, nil
}

// Status 返回登录会话的当前快照。
func (m *LoginManager) Status(sessionKey string) (*LoginSession, bool) {
	m.mu.Lock()
	defer m.mu.Unlock()
	s, ok := m.sessions[sessionKey]
	if !ok {
		return nil, false
	}
	snap := *s
	return &snap, true
}

// poll 后台轮询扫码状态，confirmed 时落 token。
func (m *LoginManager) poll(ctx context.Context, sess *LoginSession) {
	deadline := time.Now().Add(loginDeadline)
	for time.Now().Before(deadline) {
		cli := m.newClient(sess.baseURL)
		resp, err := cli.GetQRCodeStatus(ctx, sess.QRCode)
		if err != nil {
			m.setErr(sess, fmt.Sprintf("poll status: %v", err))
			return
		}

		switch resp.Status {
		case QRStatusWait, QRStatusScaned:
			m.setStatus(sess, resp.Status)
		case QRStatusScanedButRedirect:
			if resp.RedirectHost != "" {
				m.mu.Lock()
				sess.baseURL = "https://" + resp.RedirectHost
				m.mu.Unlock()
			}
		case QRStatusExpired:
			m.setErr(sess, "二维码已过期，请重新生成")
			return
		case QRStatusNeedVerifyCode, QRStatusVerifyCodeBlocked:
			// 本期不支持配对码交互（无终端 stdin），直接失败。
			m.setErr(sess, "需要配对码，当前不支持，请重试")
			return
		case QRStatusBindedRedirect:
			m.setErr(sess, "该账号已绑定，无需重复连接")
			return
		case QRStatusConfirmed:
			if resp.ILinkBotID == "" {
				m.setErr(sess, "登录确认但缺少 ilink_bot_id")
				return
			}
			m.store.Save(&Account{
				AccountID: resp.ILinkBotID,
				Token:     resp.BotToken,
				BaseURL:   resp.BaseURL,
				UserID:    resp.ILinkUserID,
			})
			m.mu.Lock()
			sess.Status = QRStatusConfirmed
			sess.AccountID = resp.ILinkBotID
			sess.UserID = resp.ILinkUserID
			m.mu.Unlock()
			return
		}

		select {
		case <-ctx.Done():
			m.setErr(sess, "登录已取消")
			return
		case <-time.After(pollInterval):
		}
	}
	m.setErr(sess, "登录超时，请重试")
}

func (m *LoginManager) setStatus(sess *LoginSession, status string) {
	m.mu.Lock()
	sess.Status = status
	m.mu.Unlock()
}

func (m *LoginManager) setErr(sess *LoginSession, msg string) {
	m.mu.Lock()
	sess.Status = QRStatusExpired
	sess.Err = msg
	m.mu.Unlock()
}
