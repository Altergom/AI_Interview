package service

import "context"

// AuthService 认证模块业务逻辑接口
type AuthService interface {
	// Register 注册新用户，返回 user_id 和 JWT token。
	Register(ctx context.Context, req RegisterRequest) (*AuthResult, error)

	// Login 邮箱密码登录，返回 user_id 和 JWT token。
	Login(ctx context.Context, req LoginRequest) (*AuthResult, error)

	// CreateGuest 创建游客账号，JWT 有效期 24h。
	CreateGuest(ctx context.Context) (*GuestResult, error)
}

type RegisterRequest struct {
	Username string
	Email    string
	Password string
}

type LoginRequest struct {
	Email    string
	Password string
}

// 改后
type AuthResult struct {
    UserID   string `json:"user_id"`
    Username string `json:"username"`
    Token    string `json:"token"`
}

type GuestResult struct {
    UserID    string `json:"user_id"`
    Token     string `json:"token"`
    ExpiresAt string `json:"expires_at"`
}
