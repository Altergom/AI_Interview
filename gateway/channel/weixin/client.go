package weixin

import (
	"bytes"
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/binary"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

// iLink 协议端点与默认参数，对照 openclaw-weixin src/api/api.ts、login-qr.ts。
const (
	defaultBaseURL = "https://ilinkai.weixin.qq.com"

	epGetBotQRCode    = "ilink/bot/get_bot_qrcode"
	epGetQRCodeStatus = "ilink/bot/get_qrcode_status"
	epGetUpdates      = "ilink/bot/getupdates"
	epSendMessage     = "ilink/bot/sendmessage"

	// defaultBotType iLink bot_type，3 为当前渠道构建（DEFAULT_ILINK_BOT_TYPE）。
	defaultBotType = "3"

	// 各类请求超时。长轮询服务端最长 hold longPollTimeout。
	longPollTimeout   = 35 * time.Second
	apiTimeout        = 15 * time.Second
	qrStatusTimeout   = 35 * time.Second
	channelVersion    = "0.1.0"
	ilinkAppID        = "bot"
	ilinkClientVersion = "65547" // 0x0001000B == 1.0.11
)

// Client iLink HTTP 客户端，单账号一个实例。token 为空表示尚未登录（仅可拉二维码）。
type Client struct {
	baseURL string
	token   string
	http    *http.Client
}

// NewClient 创建客户端。baseURL 传空用默认；token 可后续 SetToken 注入。
func NewClient(baseURL, token string) *Client {
	if baseURL == "" {
		baseURL = defaultBaseURL
	}
	return &Client{
		baseURL: strings.TrimSuffix(baseURL, "/"),
		token:   token,
		http:    &http.Client{},
	}
}

// SetToken 登录确认后注入 bot token。
func (c *Client) SetToken(token string) { c.token = token }

// randomWechatUin X-WECHAT-UIN：random uint32 → 十进制字符串 → base64。
func randomWechatUin() (string, error) {
	var b [4]byte
	if _, err := rand.Read(b[:]); err != nil {
		return "", err
	}
	u := binary.BigEndian.Uint32(b[:])
	return base64.StdEncoding.EncodeToString([]byte(strconv.FormatUint(uint64(u), 10))), nil
}

// commonHeaders GET/POST 共用头。
func (c *Client) commonHeaders(h http.Header) {
	h.Set("iLink-App-Id", ilinkAppID)
	h.Set("iLink-App-ClientVersion", ilinkClientVersion)
}

// postHeaders POST 专用头，含鉴权。
func (c *Client) postHeaders(h http.Header) error {
	h.Set("Content-Type", "application/json")
	h.Set("AuthorizationType", "ilink_bot_token")
	uin, err := randomWechatUin()
	if err != nil {
		return fmt.Errorf("weixin: gen X-WECHAT-UIN: %w", err)
	}
	h.Set("X-WECHAT-UIN", uin)
	if t := strings.TrimSpace(c.token); t != "" {
		h.Set("Authorization", "Bearer "+t)
	}
	c.commonHeaders(h)
	return nil
}

func (c *Client) baseInfo() *BaseInfo {
	return &BaseInfo{ChannelVersion: channelVersion, BotAgent: "AIInterviewGateway"}
}

// errPollTimeout 表示本次请求因客户端 deadline 超时被中断（长轮询正常情况），
// 而非外层 ctx 取消或 HTTP 错误。调用方据此决定是返回空响应重试还是上抛。
var errPollTimeout = errors.New("weixin: client poll timeout")

// doPost 发送 JSON POST，返回响应体。非 2xx 视为错误。
// 内层 deadline 超时（外层 ctx 未取消）返回 errPollTimeout。
func (c *Client) doPost(ctx context.Context, endpoint string, body any, timeout time.Duration) ([]byte, error) {
	raw, err := json.Marshal(body)
	if err != nil {
		return nil, fmt.Errorf("weixin: marshal %s: %w", endpoint, err)
	}
	reqCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()
	req, err := http.NewRequestWithContext(reqCtx, http.MethodPost, c.baseURL+"/"+endpoint, bytes.NewReader(raw))
	if err != nil {
		return nil, fmt.Errorf("weixin: new request %s: %w", endpoint, err)
	}
	if err := c.postHeaders(req.Header); err != nil {
		return nil, err
	}
	resp, err := c.http.Do(req)
	if err != nil {
		if isPollTimeout(ctx, reqCtx) {
			return nil, errPollTimeout
		}
		return nil, fmt.Errorf("weixin: post %s: %w", endpoint, err)
	}
	defer resp.Body.Close()
	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("weixin: read %s: %w", endpoint, err)
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("weixin: %s status=%d body=%s", endpoint, resp.StatusCode, string(data))
	}
	return data, nil
}

// isPollTimeout 判断错误是否源于内层 deadline 超时而非外层取消：
// reqCtx 已超时（DeadlineExceeded）但外层 ctx 仍存活。
func isPollTimeout(outer, reqCtx context.Context) bool {
	return outer.Err() == nil && errors.Is(reqCtx.Err(), context.DeadlineExceeded)
}

// doGet 发送 GET，endpoint 已含 query。返回响应体。
// 内层 deadline 超时（外层 ctx 未取消）返回 errPollTimeout。
func (c *Client) doGet(ctx context.Context, endpoint string, timeout time.Duration) ([]byte, error) {
	reqCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()
	req, err := http.NewRequestWithContext(reqCtx, http.MethodGet, c.baseURL+"/"+endpoint, nil)
	if err != nil {
		return nil, fmt.Errorf("weixin: new request %s: %w", endpoint, err)
	}
	c.commonHeaders(req.Header)
	resp, err := c.http.Do(req)
	if err != nil {
		if isPollTimeout(ctx, reqCtx) {
			return nil, errPollTimeout
		}
		return nil, fmt.Errorf("weixin: get %s: %w", endpoint, err)
	}
	defer resp.Body.Close()
	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("weixin: read %s: %w", endpoint, err)
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("weixin: %s status=%d body=%s", endpoint, resp.StatusCode, string(data))
	}
	return data, nil
}

// GetBotQRCode 拉取登录二维码。无需 token。localTokens 可为已登录账号 token（用于服务端去重），可空。
func (c *Client) GetBotQRCode(ctx context.Context, localTokens []string) (*QRCodeResp, error) {
	if localTokens == nil {
		localTokens = []string{}
	}
	ep := epGetBotQRCode + "?bot_type=" + url.QueryEscape(defaultBotType)
	data, err := c.doPost(ctx, ep, qrCodeReq{LocalTokenList: localTokens}, apiTimeout)
	if err != nil {
		return nil, err
	}
	var resp QRCodeResp
	if err := json.Unmarshal(data, &resp); err != nil {
		return nil, fmt.Errorf("weixin: unmarshal qrcode: %w", err)
	}
	return &resp, nil
}

// GetQRCodeStatus 长轮询扫码状态（GET，服务端 hold 最长 qrStatusTimeout）。
// 客户端超时（无新状态）返回 status=wait 让调用方重试，对应官方长轮询语义。
func (c *Client) GetQRCodeStatus(ctx context.Context, qrcode string) (*StatusResp, error) {
	ep := epGetQRCodeStatus + "?qrcode=" + url.QueryEscape(qrcode)
	data, err := c.doGet(ctx, ep, qrStatusTimeout)
	if err != nil {
		// 客户端长轮询超时视为继续等待，与 login-qr.ts 一致；其他错误上抛。
		if errors.Is(err, errPollTimeout) {
			return &StatusResp{Status: QRStatusWait}, nil
		}
		return nil, err
	}
	var resp StatusResp
	if err := json.Unmarshal(data, &resp); err != nil {
		return nil, fmt.Errorf("weixin: unmarshal qr status: %w", err)
	}
	return &resp, nil
}

// GetUpdates 长轮询拉取入站消息。客户端超时返回空响应（保留游标）让调用方重试。
func (c *Client) GetUpdates(ctx context.Context, updatesBuf string) (*GetUpdatesResp, error) {
	return c.getUpdatesWithTimeout(ctx, updatesBuf, longPollTimeout)
}

// getUpdatesWithTimeout 是 GetUpdates 的可注入超时版本，便于测试长轮询超时分支。
func (c *Client) getUpdatesWithTimeout(ctx context.Context, updatesBuf string, timeout time.Duration) (*GetUpdatesResp, error) {
	req := GetUpdatesReq{GetUpdatesBuf: updatesBuf, BaseInfo: c.baseInfo()}
	data, err := c.doPost(ctx, epGetUpdates, req, timeout)
	if err != nil {
		// 长轮询客户端超时是正常情况：返回空响应、保留游标，调用方续拉。
		// 其他错误（HTTP 5xx、网络故障）上抛，由 connector 触发重连退避。
		if errors.Is(err, errPollTimeout) {
			return &GetUpdatesResp{Ret: 0, GetUpdatesBuf: updatesBuf}, nil
		}
		return nil, err
	}
	var resp GetUpdatesResp
	if err := json.Unmarshal(data, &resp); err != nil {
		return nil, fmt.Errorf("weixin: unmarshal updates: %w", err)
	}
	return &resp, nil
}

// SendMessage 出站发送一条消息。
func (c *Client) SendMessage(ctx context.Context, msg *WeixinMessage) error {
	req := SendMessageReq{Msg: msg, BaseInfo: c.baseInfo()}
	_, err := c.doPost(ctx, epSendMessage, req, apiTimeout)
	return err
}
