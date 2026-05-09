package service

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"ai_interview/internal/auth"
	biz "ai_interview/internal/errors"
	"ai_interview/internal/log"
	"ai_interview/internal/storage/postgres"
)

type authService struct {
	users  *postgres.UserRepo
	jwtCfg auth.TokenConfig
}

// NewAuthService 创建 AuthService 实现。
func NewAuthService(users *postgres.UserRepo, jwtCfg auth.TokenConfig) AuthService {
	return &authService{users: users, jwtCfg: jwtCfg}
}

// Register 检查邮箱唯一 → bcrypt 加密 → 插库 → 签发 JWT。
func (s *authService) Register(ctx context.Context, req RegisterRequest) (*AuthResult, error) {
	email := strings.TrimSpace(strings.ToLower(req.Email))
	if email == "" || !strings.Contains(email, "@") {
		return nil, biz.NewMsg(biz.CodeBadRequest, "邮箱格式不正确")
	}
	if len(req.Password) < 6 {
		return nil, biz.NewMsg(biz.CodeBadRequest, "密码不能少于 6 位")
	}
	username := strings.TrimSpace(req.Username)
	if username == "" {
		return nil, biz.NewMsg(biz.CodeBadRequest, "用户名不能为空")
	}

	// 邮箱重复检查
	_, err := s.users.FindByEmail(ctx, email)
	if err == nil {
		return nil, biz.New(biz.CodeEmailRegistered)
	}
	if !errors.Is(err, postgres.ErrUserNotFound) {
		return nil, biz.Wrap(biz.CodeInternal, fmt.Errorf("check email: %w", err))
	}

	// 密码加密
	hash, err := auth.HashPassword(req.Password)
	if err != nil {
		return nil, biz.Wrap(biz.CodeInternal, err)
	}

	// 写库
	id, err := s.users.Create(ctx, postgres.UserRow{
		Email:        email,
		Username:     username,
		PasswordHash: hash,
		IsGuest:      false,
	})
	if err != nil {
		return nil, biz.Wrap(biz.CodeInternal, err)
	}

	// 签发 JWT
	token, err := auth.GenerateToken(s.jwtCfg, id, false)
	if err != nil {
		return nil, biz.Wrap(biz.CodeInternal, err)
	}

	log.Infof("[AuthService] registered user %s email=%s", id, email)
	return &AuthResult{UserID: id, Username: username, Token: token}, nil
}

// Login 留 stub，Task 5 实现。
func (s *authService) Login(ctx context.Context, req LoginRequest) (*AuthResult, error) {
	return nil, biz.NewMsg(biz.CodeInternal, "not implemented")
}

// CreateGuest 留 stub，Task 6 实现。
func (s *authService) CreateGuest(ctx context.Context) (*GuestResult, error) {
	return nil, biz.NewMsg(biz.CodeInternal, "not implemented")
}
