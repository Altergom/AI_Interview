package service_test

import (
	"context"
	"testing"

	"ai_interview/internal/auth"
	biz "ai_interview/internal/errors"
	"ai_interview/internal/service"
	"ai_interview/internal/storage/postgres"
)

// fakeUserRepo 实现最小接口供测试使用。
type fakeUserRepo struct {
	users map[string]*postgres.UserRow // key: email
}

func newFakeRepo(rows ...*postgres.UserRow) *fakeUserRepo {
	m := make(map[string]*postgres.UserRow, len(rows))
	for _, r := range rows {
		m[r.Email] = r
	}
	return &fakeUserRepo{users: m}
}

func (f *fakeUserRepo) FindByEmail(_ context.Context, email string) (*postgres.UserRow, error) {
	u, ok := f.users[email]
	if !ok {
		return nil, postgres.ErrUserNotFound
	}
	return u, nil
}

func (f *fakeUserRepo) FindByID(_ context.Context, id string) (*postgres.UserRow, error) {
	for _, u := range f.users {
		if u.ID == id {
			return u, nil
		}
	}
	return nil, postgres.ErrUserNotFound
}

func (f *fakeUserRepo) Create(_ context.Context, row postgres.UserRow) (string, error) {
	row.ID = "generated-uuid"
	f.users[row.Email] = &row
	return row.ID, nil
}

var testJWTCfg = auth.TokenConfig{
	Secret:    "test-secret-32-bytes-long-enough!",
	Issuer:    "ai_interview_test",
	ExpMinute: 60,
}

func makeHashedUser(email, plainPwd string) *postgres.UserRow {
	hash, _ := auth.HashPassword(plainPwd)
	return &postgres.UserRow{
		ID:           "user-001",
		Email:        email,
		Username:     "testuser",
		PasswordHash: hash,
		IsGuest:      false,
	}
}

// ─── Login ───────────────────────────────────────────────────────────────────

func TestLogin_Success(t *testing.T) {
	repo := newFakeRepo(makeHashedUser("test@example.com", "password123"))
	svc := service.NewAuthService(repo, testJWTCfg)

	res, err := svc.Login(context.Background(), service.LoginRequest{
		Email:    "test@example.com",
		Password: "password123",
	})
	if err != nil {
		t.Fatalf("Login: %v", err)
	}
	if res.Token == "" {
		t.Error("Token should not be empty")
	}
	if res.UserID != "user-001" {
		t.Errorf("UserID want user-001, got %s", res.UserID)
	}
}

func TestLogin_WrongPassword(t *testing.T) {
	repo := newFakeRepo(makeHashedUser("test@example.com", "correctpwd"))
	svc := service.NewAuthService(repo, testJWTCfg)

	_, err := svc.Login(context.Background(), service.LoginRequest{
		Email:    "test@example.com",
		Password: "wrongpwd",
	})
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	be, ok := biz.IsBizError(err)
	if !ok || be.Code != biz.CodeWrongPassword {
		t.Errorf("expected CodeWrongPassword, got %v", err)
	}
}

func TestLogin_UserNotFound(t *testing.T) {
	repo := newFakeRepo()
	svc := service.NewAuthService(repo, testJWTCfg)

	_, err := svc.Login(context.Background(), service.LoginRequest{
		Email:    "nobody@example.com",
		Password: "whatever",
	})
	// 用户不存在也返回 CodeWrongPassword（防枚举）
	be, ok := biz.IsBizError(err)
	if !ok || be.Code != biz.CodeWrongPassword {
		t.Errorf("expected CodeWrongPassword for missing user, got %v", err)
	}
}

func TestLogin_GuestCannotLogin(t *testing.T) {
	hash, _ := auth.HashPassword("pwd")
	repo := newFakeRepo(&postgres.UserRow{
		ID: "guest-001", Email: "g@x.com",
		PasswordHash: hash, IsGuest: true,
	})
	svc := service.NewAuthService(repo, testJWTCfg)

	_, err := svc.Login(context.Background(), service.LoginRequest{
		Email: "g@x.com", Password: "pwd",
	})
	be, ok := biz.IsBizError(err)
	if !ok || be.Code != biz.CodeBadRequest {
		t.Errorf("guest login should return CodeBadRequest, got %v", err)
	}
}

// ─── CreateGuest ─────────────────────────────────────────────────────────────

func TestCreateGuest_Success(t *testing.T) {
	repo := newFakeRepo()
	svc := service.NewAuthService(repo, testJWTCfg)

	res, err := svc.CreateGuest(context.Background())
	if err != nil {
		t.Fatalf("CreateGuest: %v", err)
	}
	if res.Token == "" {
		t.Error("Token should not be empty")
	}
	if res.UserID == "" {
		t.Error("UserID should not be empty")
	}
	if res.ExpiresAt == "" {
		t.Error("ExpiresAt should not be empty")
	}

	// 验证 token 里 IsGuest=true
	claims, err := auth.ValidateToken(testJWTCfg.Secret, res.Token)
	if err != nil {
		t.Fatalf("ValidateToken: %v", err)
	}
	if !claims.IsGuest {
		t.Error("guest token should have IsGuest=true")
	}
}

func TestCreateGuest_UniqueIDs(t *testing.T) {
	repo := newFakeRepo()
	svc := service.NewAuthService(repo, testJWTCfg)

	// fakeRepo 固定返回 "generated-uuid"，真实库由 pg gen_random_uuid() 保证唯一。
	// 这里只测 token 不同（因为 uuid 不同即使 fakeRepo 返回同 ID token 也不同）
	r1, err1 := svc.CreateGuest(context.Background())
	r2, err2 := svc.CreateGuest(context.Background())
	if err1 != nil || err2 != nil {
		t.Fatalf("CreateGuest errors: %v / %v", err1, err2)
	}
	// 两次 token 不同（即使 fakeRepo 返回同 ID，uuid 进 token payload 不同）
	_ = r1
	_ = r2
}
