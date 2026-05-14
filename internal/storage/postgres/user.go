package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"
)

// UserRepository 定义 users 表读写接口，便于 service 层测试 mock。
type UserRepository interface {
	FindByEmail(ctx context.Context, email string) (*UserRow, error)
	FindByID(ctx context.Context, id string) (*UserRow, error)
	Create(ctx context.Context, u UserRow) (string, error)
}

// UserRow 对应 users 表一行。
type UserRow struct {
	ID           string
	Email        string
	Username     string
	PasswordHash string
	IsGuest      bool
	CreatedAt    time.Time
}

// ErrUserNotFound 用户不存在哨兵错误。
var ErrUserNotFound = errors.New("user not found")

// UserRepo 封装 users 表的读写操作。
type UserRepo struct {
	db *sql.DB
}

// NewUserRepo 创建 UserRepo。
func NewUserRepo(db *sql.DB) *UserRepo {
	return &UserRepo{db: db}
}

// FindByEmail 按邮箱查找用户，不存在返回 ErrUserNotFound。
func (r *UserRepo) FindByEmail(ctx context.Context, email string) (*UserRow, error) {
	row := r.db.QueryRowContext(ctx,
		`SELECT id, email, username, password_hash, is_guest, created_at
		   FROM users WHERE email = $1`, email)

	var u UserRow
	err := row.Scan(&u.ID, &u.Email, &u.Username, &u.PasswordHash, &u.IsGuest, &u.CreatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrUserNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("find user by email: %w", err)
	}
	return &u, nil
}

// FindByID 按 UUID 查找用户，不存在返回 ErrUserNotFound。
func (r *UserRepo) FindByID(ctx context.Context, id string) (*UserRow, error) {
	row := r.db.QueryRowContext(ctx,
		`SELECT id, email, username, password_hash, is_guest, created_at
		   FROM users WHERE id = $1`, id)

	var u UserRow
	err := row.Scan(&u.ID, &u.Email, &u.Username, &u.PasswordHash, &u.IsGuest, &u.CreatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrUserNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("find user by id: %w", err)
	}
	return &u, nil
}

// Create 插入新用户，返回数据库生成的 UUID。
func (r *UserRepo) Create(ctx context.Context, u UserRow) (string, error) {
	var id string
	err := r.db.QueryRowContext(ctx,
		`INSERT INTO users (email, username, password_hash, is_guest)
		 VALUES ($1, $2, $3, $4)
		 RETURNING id`,
		u.Email, u.Username, u.PasswordHash, u.IsGuest,
	).Scan(&id)
	if err != nil {
		return "", fmt.Errorf("create user: %w", err)
	}
	return id, nil
}
