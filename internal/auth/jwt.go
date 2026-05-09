package auth

import (
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// Claims 是 JWT payload 的业务扩展。
type Claims struct {
	UserID  string `json:"uid"`
	IsGuest bool   `json:"is_guest"`
	jwt.RegisteredClaims
}

// TokenConfig JWT 签发参数，由调用方从 config 注入，不在包内读全局变量。
type TokenConfig struct {
	Secret    string
	Issuer    string
	ExpMinute int // 普通用户过期时间（分钟）
}

// GenerateToken 签发 JWT。
// 若 isGuest=true，有效期固定 24h，忽略 cfg.ExpMinute。
func GenerateToken(cfg TokenConfig, userID string, isGuest bool) (string, error) {
	if cfg.Secret == "" {
		return "", errors.New("JWT_SECRET is not set")
	}

	expiry := time.Duration(cfg.ExpMinute) * time.Minute
	if isGuest {
		expiry = 24 * time.Hour
	}

	now := time.Now()
	claims := Claims{
		UserID:  userID,
		IsGuest: isGuest,
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    cfg.Issuer,
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(expiry)),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signed, err := token.SignedString([]byte(cfg.Secret))
	if err != nil {
		return "", fmt.Errorf("sign jwt: %w", err)
	}
	return signed, nil
}

// ValidateToken 解析并验证 JWT，返回业务 Claims。
func ValidateToken(secret, tokenStr string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenStr, &Claims{}, func(t *jwt.Token) (any, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
		}
		return []byte(secret), nil
	})
	if err != nil {
		return nil, fmt.Errorf("parse jwt: %w", err)
	}

	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		return nil, errors.New("invalid jwt claims")
	}
	return claims, nil
}
