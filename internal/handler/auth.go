package handler

import (
	"context"

	"github.com/cloudwego/hertz/pkg/app"

	"ai_interview/internal/service"
)

type authHandler struct {
	svc service.AuthService
}

// Register POST /v1/auth/register
func (h *authHandler) Register(ctx context.Context, c *app.RequestContext) {
	panic("not implemented")
}

// Login POST /v1/auth/login
func (h *authHandler) Login(ctx context.Context, c *app.RequestContext) {
	panic("not implemented")
}

// Guest POST /v1/auth/guest
// 生成临时游客账号，JWT 有效期 24h。
func (h *authHandler) Guest(ctx context.Context, c *app.RequestContext) {
	panic("not implemented")
}
