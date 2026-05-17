package postgres

import (
	"context"
	"errors"
	"fmt"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// UserRepository defines users table operations for auth service.
type UserRepository interface {
	FindByEmail(ctx context.Context, email string) (*UserRow, error)
	FindByID(ctx context.Context, id string) (*UserRow, error)
	Create(ctx context.Context, u UserRow) (string, error)
}

// UserRow is the storage row shape consumed by the auth service.
type UserRow struct {
	ID           string
	Email        string
	Username     string
	PasswordHash string
	IsGuest      bool
	CreatedAt    time.Time
}

// ErrUserNotFound indicates a missing user row.
var ErrUserNotFound = errors.New("user not found")

// UserRepo wraps users table reads and writes.
type UserRepo struct {
	db *gorm.DB
}

// NewUserRepo creates a UserRepo.
func NewUserRepo(db *gorm.DB) *UserRepo {
	return &UserRepo{db: db}
}

// FindByEmail finds a user by email, returning ErrUserNotFound when missing.
func (r *UserRepo) FindByEmail(ctx context.Context, email string) (*UserRow, error) {
	var model UserModel
	err := r.db.WithContext(ctx).Where(&UserModel{Email: email}).First(&model).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrUserNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("find user by email: %w", err)
	}
	return userModelToRow(model), nil
}

// FindByID finds a user by UUID, returning ErrUserNotFound when missing.
func (r *UserRepo) FindByID(ctx context.Context, id string) (*UserRow, error) {
	var model UserModel
	err := r.db.WithContext(ctx).Where(&UserModel{ID: id}).First(&model).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrUserNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("find user by id: %w", err)
	}
	return userModelToRow(model), nil
}

// Create inserts a new user and returns the database-generated UUID.
func (r *UserRepo) Create(ctx context.Context, u UserRow) (string, error) {
	model := UserModel{
		Email:        u.Email,
		Username:     u.Username,
		PasswordHash: u.PasswordHash,
		IsGuest:      u.IsGuest,
	}
	err := r.db.WithContext(ctx).
		Clauses(clause.Returning{Columns: []clause.Column{{Name: "id"}}}).
		Create(&model).Error
	if err != nil {
		return "", fmt.Errorf("create user: %w", err)
	}
	return model.ID, nil
}

func userModelToRow(model UserModel) *UserRow {
	return &UserRow{
		ID:           model.ID,
		Email:        model.Email,
		Username:     model.Username,
		PasswordHash: model.PasswordHash,
		IsGuest:      model.IsGuest,
		CreatedAt:    model.CreatedAt,
	}
}
