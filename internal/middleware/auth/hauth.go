package auth

import (
	"context"
	"net/http"
	"strings"

	"github.com/cloudwego/hertz/pkg/app"

	"ai_interview/internal/auth"
	biz "ai_interview/internal/errors"
)

// Context key 常量，供 handler/service 层读取。
// 与 ratelimit 中间件约定的 "user_id" 保持一致。
const (
	CtxKeyUserID  = "user_id"
	CtxKeyIsGuest = "is_guest"
)

// authErrResp 401 响应体，与 handler.Result[T] 格式保持一致，避免 import cycle。
type authErrResp struct {
	Success bool         `json:"success"`
	Data    any          `json:"data"`
	Error   *authErrBody `json:"error"`
}

type authErrBody struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

func rejectUnauthorized(ctx context.Context, c *app.RequestContext, msg string) {
	c.JSON(http.StatusUnauthorized, authErrResp{
		Success: false,
		Data:    nil,
		Error:   &authErrBody{Code: int(biz.CodeUnauthorized), Message: msg},
	})
	c.Abort()
}

// HAuth 返回 JWT 鉴权中间件。
//
// 行为：
//   - 读取 Authorization: Bearer <token>
//   - 解析验证 JWT，失败返回 401
//   - 成功将 user_id(string) 和 is_guest(bool) 注入 Hertz context
//   - 游客 token 同样通过（is_guest=true）
//
// 用法：
//
//	protected := v1.Group("/resume", authmw.HAuth(jwtSecret))
func HAuth(secret string) app.HandlerFunc {
	return func(ctx context.Context, c *app.RequestContext) {
		token := extractBearer(c)
		if token == "" {
			rejectUnauthorized(ctx, c, "missing or malformed authorization header")
			return
		}

		claims, err := auth.ValidateToken(secret, token)
		if err != nil {
			rejectUnauthorized(ctx, c, "invalid or expired token")
			return
		}

		c.Set(CtxKeyUserID, claims.UserID)
		c.Set(CtxKeyIsGuest, claims.IsGuest)
		c.Next(ctx)
	}
}

// extractBearer 从 Authorization 头提取 Bearer token。
func extractBearer(c *app.RequestContext) string {
	header := string(c.GetHeader("Authorization"))
	if !strings.HasPrefix(header, "Bearer ") {
		return ""
	}
	token := strings.TrimSpace(strings.TrimPrefix(header, "Bearer "))
	if token == "" {
		return ""
	}
	return token
}

// GetUserID 从 Hertz context 中读取当前用户 ID（由 HAuth 注入）。
func GetUserID(c *app.RequestContext) string {
	v, _ := c.Get(CtxKeyUserID)
	id, _ := v.(string)
	return id
}

// GetIsGuest 从 Hertz context 中读取当前用户是否为游客（由 HAuth 注入）。
func GetIsGuest(c *app.RequestContext) bool {
	v, _ := c.Get(CtxKeyIsGuest)
	b, _ := v.(bool)
	return b
}
