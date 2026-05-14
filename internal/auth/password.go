package auth

import (
	"errors"
	"fmt"

	"golang.org/x/crypto/bcrypt"
)

// bcrypt cost：12 在现代硬件约 200-300ms，兼顾安全性与用户体验。
// 生产如需调高可改此常量并重新编译；已有密码不受影响（bcrypt 自带版本信息）。
const bcryptCost = 12

// ErrWrongPassword 密码不匹配时返回，调用方可据此返回 401 而不暴露"用户不存在"。
var ErrWrongPassword = errors.New("wrong password")

// HashPassword 对明文密码做 bcrypt hash，返回可直接存库的字符串。
func HashPassword(plain string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(plain), bcryptCost)
	if err != nil {
		return "", fmt.Errorf("hash password: %w", err)
	}
	return string(hash), nil
}

// ComparePassword 验证明文密码与存库 hash 是否匹配。
// 不匹配返回 ErrWrongPassword，其他错误透传。
func ComparePassword(hash, plain string) error {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(plain))
	if errors.Is(err, bcrypt.ErrMismatchedHashAndPassword) {
		return ErrWrongPassword
	}
	if err != nil {
		return fmt.Errorf("compare password: %w", err)
	}
	return nil
}
