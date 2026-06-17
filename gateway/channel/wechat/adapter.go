package wechat

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"gateway/inbound"
)

// Config 微信服务号 Webhook 配置，全部从环境变量注入。
type Config struct {
	Token          string // 微信后台配置的 Token，用于签名验证
	EncodingAESKey string // 43 位消息加密密钥（Base64）
	ReceiveID      string // 服务号原始 ID（gh_xxxx）
	CallbackURL    string // 微信客服消息回调基址（主动回复用，含 access_token）
}

// Adapter 微信服务号渠道适配器。
type Adapter struct {
	cfg Config
}

func New(cfg Config) *Adapter {
	return &Adapter{cfg: cfg}
}

func (a *Adapter) Name() string { return "wechat" }

// VerifySignature 验证微信推送请求签名。
// headers 里需要包含从 query 参数提取的 timestamp/nonce/msg_signature/encrypt。
func (a *Adapter) VerifySignature(_ context.Context, headers map[string]string, _ []byte) error {
	timestamp := headers["timestamp"]
	nonce := headers["nonce"]
	encrypt := headers["encrypt"]
	signature := headers["msg_signature"]
	if signature == "" {
		signature = headers["signature"]
	}
	if !verifySignature(a.cfg.Token, timestamp, nonce, encrypt, signature) {
		return fmt.Errorf("wechat: signature mismatch")
	}
	return nil
}

// Parse 解析微信推送消息，支持加密模式和明文模式。
func (a *Adapter) Parse(_ context.Context, body []byte) (*inbound.InboundEvent, error) {
	msg, err := parseXMLMessage(body)
	if err != nil {
		return nil, err
	}

	if msg.Encrypt != "" {
		plaintext, err := decryptMessage(a.cfg.EncodingAESKey, a.cfg.ReceiveID, msg.Encrypt)
		if err != nil {
			return nil, fmt.Errorf("wechat: decrypt failed: %w", err)
		}
		inner, err := parseXMLMessage([]byte(plaintext))
		if err != nil {
			return nil, fmt.Errorf("wechat: parse decrypted xml: %w", err)
		}
		msg = inner
	}

	msgType := msg.MsgType
	if msgType == "" {
		msgType = "text"
	}

	return &inbound.InboundEvent{
		Channel:    "wechat",
		AccountID:  msg.ToUserName,
		PeerID:     msg.FromUserName,
		PeerType:   "user",
		MessageID:  msg.MsgId,
		Payload:    inbound.EventPayload{Type: msgType, Content: msg.Content},
		Raw:        body,
		ReceivedAt: time.Unix(msg.CreateTime, 0),
	}, nil
}

// Ack 普通消息 ack，微信接受空字符串响应。
// GET 验证场景通过 HandleVerify 单独处理。
func (a *Adapter) Ack(_ context.Context, _ []byte) (any, error) {
	return "", nil
}

// Send 通过微信客服消息接口主动回复用户。
func (a *Adapter) Send(_ context.Context, _ string, _ string) error {
	if a.cfg.CallbackURL == "" {
		return fmt.Errorf("wechat: CallbackURL not configured")
	}
	// TODO: POST /cgi-bin/message/custom/send，需要 access_token 管理
	return nil
}

// HandleVerify 处理微信服务器 URL 验证（GET echostr 场景）。
// 验证签名后解密 echostr 并原样返回，由 webhook handler 调用。
func (a *Adapter) HandleVerify(timestamp, nonce, signature, echostr string) (string, error) {
	if !verifySignature(a.cfg.Token, timestamp, nonce, echostr, signature) {
		return "", fmt.Errorf("wechat: verify signature failed")
	}
	plain, err := decryptMessage(a.cfg.EncodingAESKey, a.cfg.ReceiveID, echostr)
	if err != nil {
		return "", fmt.Errorf("wechat: decrypt echostr: %w", err)
	}
	return plain, nil
}

// BuildEncryptedReply 构建加密模式的 XML 回复，用于同步回复微信消息。
func (a *Adapter) BuildEncryptedReply(toUser, replyText, nonce string) (string, error) {
	timestamp := strconv.FormatInt(time.Now().Unix(), 10)

	innerXML := buildXMLReply(toUser, a.cfg.ReceiveID, replyText, "", "", timestamp, nonce)
	encrypt, err := encryptMessage(a.cfg.EncodingAESKey, a.cfg.ReceiveID, innerXML)
	if err != nil {
		return "", fmt.Errorf("wechat: encrypt reply: %w", err)
	}

	sig := buildSignature(a.cfg.Token, timestamp, nonce, encrypt)
	return buildXMLReply("", "", "", encrypt, sig, timestamp, nonce), nil
}
