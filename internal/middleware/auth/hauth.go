package auth

import (
	"context"
	"net/http"

	"github.com/cloudwego/hertz/pkg/app"

	"ai_interview/internal/auth"
	"ai_interview/internal/utils/hertzx"
	biz "ai_interview/internal/utils/respx"
)

// ctxKey 防止 context key 冲突的私有类型。
type ctxKey string

const ctxUserIDKey ctxKey = "user_id"

// WithUserID 将 userID 注入标准 context（供 WebSocket handler 向 service 层传递）。
func WithUserID(ctx context.Context, userID string) context.Context {
	return context.WithValue(ctx, ctxUserIDKey, userID)
}

// UserIDFromContext 从标准 context 中读取 userID（由 WithUserID 注入）。
func UserIDFromContext(ctx context.Context) string {
	v, _ := ctx.Value(ctxUserIDKey).(string)
	return v
}

// Context key 常量，供 handler/service 层读取。
// 与 ratelimit 中间件约定的 "user_id" 保持一致。
const (
	CtxKeyUserID  = "user_id"
	CtxKeyIsGuest = "is_guest"
)

func rejectUnauthorized(ctx context.Context, c *app.RequestContext, msg string) {
	_ = msg
	c.JSON(http.StatusOK, biz.Fail(biz.CodeUnauthorized))
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
		token := hertzx.BearerToken(c)
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
