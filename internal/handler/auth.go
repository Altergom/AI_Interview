package handler

import (
	"context"
	"net/http"

	"github.com/cloudwego/hertz/pkg/app"

	biz "ai_interview/internal/errors"
	"ai_interview/internal/service"
)

type authHandler struct {
	svc service.AuthService
}

type registerReq struct {
	Username string `json:"username"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

type loginReq struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

// Register POST /v1/auth/register
func (h *authHandler) Register(ctx context.Context, c *app.RequestContext) {
	var req registerReq
	if err := c.BindJSON(&req); err != nil {
		Fail(ctx, c, http.StatusBadRequest, biz.CodeBadRequest, "invalid request body")
		return
	}

	result, err := h.svc.Register(ctx, service.RegisterRequest{
		Username: req.Username,
		Email:    req.Email,
		Password: req.Password,
	})
	if err != nil {
		HandleErr(ctx, c, err)
		return
	}

	OK(ctx, c, result)
}

// Login POST /v1/auth/login
func (h *authHandler) Login(ctx context.Context, c *app.RequestContext) {
	var req loginReq
	if err := c.BindJSON(&req); err != nil {
		Fail(ctx, c, http.StatusBadRequest, biz.CodeBadRequest, "invalid request body")
		return
	}

	result, err := h.svc.Login(ctx, service.LoginRequest{
		Email:    req.Email,
		Password: req.Password,
	})
	if err != nil {
		HandleErr(ctx, c, err)
		return
	}

	OK(ctx, c, result)
}

// Guest POST /v1/auth/guest
// 无需请求体，直接生成游客账号，返回 24h JWT。
func (h *authHandler) Guest(ctx context.Context, c *app.RequestContext) {
	result, err := h.svc.CreateGuest(ctx)
	if err != nil {
		HandleErr(ctx, c, err)
		return
	}
	OK(ctx, c, result)
}
