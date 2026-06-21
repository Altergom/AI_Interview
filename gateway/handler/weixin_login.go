package handler

import (
	"context"
	"net/http"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/google/uuid"

	"gateway/channel/weixin"
)

// WeixinLoginHandler 微信扫码登录接口。后端代轮询 iLink，前端只负责把二维码 URL 渲染成图。
type WeixinLoginHandler struct {
	mgr *weixin.LoginManager
}

func NewWeixinLoginHandler(mgr *weixin.LoginManager) *WeixinLoginHandler {
	return &WeixinLoginHandler{mgr: mgr}
}

// Start POST /v1/gateway/weixin/login
// 发起登录，返回 session_key（后续查状态用）和二维码 URL（前端渲染）。
func (h *WeixinLoginHandler) Start(ctx context.Context, c *app.RequestContext) {
	sessionKey := uuid.NewString()
	sess, err := h.mgr.Start(ctx, sessionKey)
	if err != nil {
		status, resp := fail(http.StatusBadGateway, 502, "发起登录失败: "+err.Error())
		c.JSON(status, resp)
		return
	}
	status, resp := ok(map[string]any{
		"session_key": sessionKey,
		"qrcode":      sess.QRCode,
		"qrcode_url":  sess.QRCodeURL,
	})
	c.JSON(status, resp)
}

// Status GET /v1/gateway/weixin/login/status?session_key=X
// 查询扫码状态。confirmed 时返回 account_id。
func (h *WeixinLoginHandler) Status(_ context.Context, c *app.RequestContext) {
	sessionKey := string(c.Query("session_key"))
	if sessionKey == "" {
		status, resp := fail(http.StatusBadRequest, 400, "session_key is required")
		c.JSON(status, resp)
		return
	}
	sess, found := h.mgr.Status(sessionKey)
	if !found {
		status, resp := fail(http.StatusNotFound, 404, "登录会话不存在或已结束")
		c.JSON(status, resp)
		return
	}
	data := map[string]any{"status": sess.Status}
	if sess.Status == weixin.QRStatusConfirmed {
		data["account_id"] = sess.AccountID
		data["user_id"] = sess.UserID
	}
	if sess.Err != "" {
		data["error"] = sess.Err
	}
	status, resp := ok(data)
	c.JSON(status, resp)
}
