package handler

import (
	"context"
	"io"
	"net/http"

	"github.com/cloudwego/hertz/pkg/app"

	"gateway/channel"
	"gateway/channel/wechat"
)

// WebhookHandler 处理 /webhook/:channel 的入站请求。
type WebhookHandler struct {
	adapters map[string]channel.ChannelAdapter
	wechat   *wechat.Adapter // 单独持有，处理 GET 验证场景
}

func NewWebhookHandler(adapters ...channel.ChannelAdapter) *WebhookHandler {
	m := make(map[string]channel.ChannelAdapter, len(adapters))
	var wx *wechat.Adapter
	for _, a := range adapters {
		m[a.Name()] = a
		if w, ok := a.(*wechat.Adapter); ok {
			wx = w
		}
	}
	return &WebhookHandler{adapters: m, wechat: wx}
}

// Handle 路由到对应渠道适配器。
// 微信 GET 请求（服务器验证）单独处理，其余 POST 走通用流程。
func (h *WebhookHandler) Handle(ctx context.Context, c *app.RequestContext) {
	ch := c.Param("channel")
	adapter, ok := h.adapters[ch]
	if !ok {
		status, resp := fail(http.StatusBadRequest, 400, "unknown channel: "+ch)
		c.JSON(status, resp)
		return
	}

	// 微信服务器 URL 验证（GET echostr）
	if ch == "wechat" && string(c.Method()) == http.MethodGet && h.wechat != nil {
		echostr := string(c.Query("echostr"))
		timestamp := string(c.Query("timestamp"))
		nonce := string(c.Query("nonce"))
		signature := string(c.Query("signature"))
		plain, err := h.wechat.HandleVerify(timestamp, nonce, signature, echostr)
		if err != nil {
			c.String(http.StatusUnauthorized, "invalid signature")
			return
		}
		c.String(http.StatusOK, plain)
		return
	}

	body, err := io.ReadAll(c.RequestBodyStream())
	if err != nil {
		status, resp := fail(http.StatusBadRequest, 400, "failed to read body")
		c.JSON(status, resp)
		return
	}

	// 把 HTTP header + query 参数合并进 map，适配器按需取用
	meta := extractMeta(c, body)
	if err := adapter.VerifySignature(ctx, meta, body); err != nil {
		status, resp := fail(http.StatusUnauthorized, 401, "signature verification failed")
		c.JSON(status, resp)
		return
	}

	ack, err := adapter.Ack(ctx, body)
	if err != nil {
		status, resp := fail(http.StatusInternalServerError, 500, "ack failed")
		c.JSON(status, resp)
		return
	}

	// TODO: 异步投入 MQ 处理
	c.JSON(http.StatusOK, ack)
}

// extractMeta 把 query 参数和部分 header 合并为 map，统一传给适配器。
// 微信用 query 参数传签名字段，这里统一提取。
func extractMeta(c *app.RequestContext, body []byte) map[string]string {
	m := map[string]string{
		"timestamp":     string(c.Query("timestamp")),
		"nonce":         string(c.Query("nonce")),
		"signature":     string(c.Query("signature")),
		"msg_signature": string(c.Query("msg_signature")),
	}

	// 从 body 里提取 encrypt 字段（微信加密模式）
	// 先尝试 XML 解析拿 Encrypt，失败忽略
	if len(body) > 0 {
		const tag = "<Encrypt><![CDATA["
		if idx := indexOf(body, []byte(tag)); idx >= 0 {
			start := idx + len(tag)
			end := indexOf(body[start:], []byte("]]></Encrypt>"))
			if end >= 0 {
				m["encrypt"] = string(body[start : start+end])
			}
		}
	}

	return m
}

func indexOf(haystack, needle []byte) int {
	for i := 0; i <= len(haystack)-len(needle); i++ {
		if string(haystack[i:i+len(needle)]) == string(needle) {
			return i
		}
	}
	return -1
}
