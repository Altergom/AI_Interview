package auth_test

import (
	"errors"
	"testing"

	"ai_interview/internal/auth"
)

func TestHashAndCompare(t *testing.T) {
	hash, err := auth.HashPassword("mySecret123!")
	if err != nil {
		t.Fatalf("HashPassword: %v", err)
	}
	if hash == "mySecret123!" {
		t.Error("hash should not equal plaintext")
	}

	if err := auth.ComparePassword(hash, "mySecret123!"); err != nil {
		t.Errorf("ComparePassword correct: %v", err)
	}
}

func TestComparePassword_Wrong(t *testing.T) {
	hash, _ := auth.HashPassword("correct")
	err := auth.ComparePassword(hash, "wrong")
	if !errors.Is(err, auth.ErrWrongPassword) {
		t.Errorf("expected ErrWrongPassword, got %v", err)
	}
}

func TestHashPassword_DifferentEachTime(t *testing.T) {
	h1, _ := auth.HashPassword("same")
	h2, _ := auth.HashPassword("same")
	// bcrypt 每次 salt 不同，hash 结果不同
	if h1 == h2 {
		t.Error("bcrypt should produce different hashes for same input")
	}
}
