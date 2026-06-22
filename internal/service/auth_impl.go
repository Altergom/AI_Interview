package service

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"ai_interview/internal/auth"
	"ai_interview/internal/log"
	"ai_interview/internal/storage/postgres"
	"ai_interview/internal/utils/uuidx"
	biz "ai_interview/internal/utils/respx"
)

type authService struct {
	users  postgres.UserRepository
	jwtCfg auth.TokenConfig
}

// NewAuthService 创建 AuthService 实现。
func NewAuthService(users postgres.UserRepository, jwtCfg auth.TokenConfig) AuthService {
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

// Login 查用户 → 验密码 → 签发 JWT。
func (s *authService) Login(ctx context.Context, req LoginRequest) (*AuthResult, error) {
	email := strings.TrimSpace(strings.ToLower(req.Email))
	if email == "" {
		return nil, biz.NewMsg(biz.CodeBadRequest, "邮箱不能为空")
	}

	// 查用户
	user, err := s.users.FindByEmail(ctx, email)
	if errors.Is(err, postgres.ErrUserNotFound) {
		// 不区分"用户不存在"和"密码错误"，防止用户枚举
		return nil, biz.New(biz.CodeWrongPassword)
	}
	if err != nil {
		return nil, biz.Wrap(biz.CodeInternal, fmt.Errorf("find user: %w", err))
	}

	// 游客账号不允许密码登录
	if user.IsGuest {
		return nil, biz.NewMsg(biz.CodeBadRequest, "游客账号不支持密码登录")
	}

	// 验密码
	if err := auth.ComparePassword(user.PasswordHash, req.Password); err != nil {
		return nil, biz.New(biz.CodeWrongPassword)
	}

	// 签发 JWT
	token, err := auth.GenerateToken(s.jwtCfg, user.ID, false)
	if err != nil {
		return nil, biz.Wrap(biz.CodeInternal, err)
	}

	log.Infof("[AuthService] login user %s email=%s", user.ID, email)
	return &AuthResult{UserID: user.ID, Username: user.Username, Token: token}, nil
}

// CreateGuest 生成游客账号：guest_{uuid8} 前缀 ID + is_guest=true + 24h JWT。
func (s *authService) CreateGuest(ctx context.Context) (*GuestResult, error) {
	// 用 uuid 生成唯一短标识，前 8 位作为展示名后缀
	shortID := uuidx.NewShort(8)
	username := "guest_" + shortID
	// 邮箱占位，保证 NOT NULL UNIQUE 约束
	email := username + "@guest.local"

	id, err := s.users.Create(ctx, postgres.UserRow{
		Email:        email,
		Username:     username,
		PasswordHash: "", // 游客无密码
		IsGuest:      true,
	})
	if err != nil {
		return nil, biz.Wrap(biz.CodeInternal, err)
	}

	token, err := auth.GenerateToken(s.jwtCfg, id, true) // isGuest=true → 24h
	if err != nil {
		return nil, biz.Wrap(biz.CodeInternal, err)
	}

	expiresAt := time.Now().Add(24 * time.Hour).UTC().Format(time.RFC3339)
	log.Infof("[AuthService] created guest %s id=%s", username, id)
	return &GuestResult{UserID: id, Token: token, ExpiresAt: expiresAt}, nil
}
