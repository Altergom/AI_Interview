package auth

import (
	"context"
	"net/http"
	"strings"

	"github.com/cloudwego/hertz/pkg/app"

	"ai_interview/internal/auth"
	biz "ai_interview/internal/errors"
	"ai_interview/internal/handler"
)

// Context key 常量，供 handler/service 层读取。
// 与 ratelimit 中间件约定的 "user_id" 保持一致。
const (
	CtxKeyUserID  = "user_id"
	CtxKeyIsGuest = "is_guest"
)

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
			handler.Fail(ctx, c, http.StatusUnauthorized, biz.CodeUnauthorized, "missing or malformed authorization header")
			c.Abort()
			return
		}

		claims, err := auth.ValidateToken(secret, token)
		if err != nil {
			handler.Fail(ctx, c, http.StatusUnauthorized, biz.CodeUnauthorized, "invalid or expired token")
			c.Abort()
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
