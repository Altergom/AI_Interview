package auth_test

import (
	"testing"
	"time"

	"ai_interview/internal/auth"
)

var testCfg = auth.TokenConfig{
	Secret:    "test-secret-32-bytes-long-enough!",
	Issuer:    "ai_interview",
	ExpMinute: 60,
}

func TestGenerateAndValidate(t *testing.T) {
	token, err := auth.GenerateToken(testCfg, "user-123", false)
	if err != nil {
		t.Fatalf("GenerateToken: %v", err)
	}

	claims, err := auth.ValidateToken(testCfg.Secret, token)
	if err != nil {
		t.Fatalf("ValidateToken: %v", err)
	}
	if claims.UserID != "user-123" {
		t.Errorf("UserID want user-123, got %s", claims.UserID)
	}
	if claims.IsGuest {
		t.Error("IsGuest should be false")
	}
}

func TestGuestToken24h(t *testing.T) {
	token, err := auth.GenerateToken(testCfg, "guest_abc12345", true)
	if err != nil {
		t.Fatalf("GenerateToken guest: %v", err)
	}

	claims, err := auth.ValidateToken(testCfg.Secret, token)
	if err != nil {
		t.Fatalf("ValidateToken guest: %v", err)
	}
	if !claims.IsGuest {
		t.Error("IsGuest should be true")
	}
	// 有效期应在 23h59m ~ 24h 之间
	remaining := time.Until(claims.ExpiresAt.Time)
	if remaining < 23*time.Hour+59*time.Minute || remaining > 24*time.Hour+time.Minute {
		t.Errorf("guest token expiry unexpected: %v", remaining)
	}
}

func TestValidateToken_WrongSecret(t *testing.T) {
	token, _ := auth.GenerateToken(testCfg, "user-456", false)
	_, err := auth.ValidateToken("wrong-secret", token)
	if err == nil {
		t.Error("expected error with wrong secret, got nil")
	}
}

func TestGenerateToken_EmptySecret(t *testing.T) {
	cfg := auth.TokenConfig{Secret: "", Issuer: "test", ExpMinute: 60}
	_, err := auth.GenerateToken(cfg, "user-789", false)
	if err == nil {
		t.Error("expected error with empty secret, got nil")
	}
}
