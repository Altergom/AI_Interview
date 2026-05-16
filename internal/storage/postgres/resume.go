package postgres

import (
	"context"
	"errors"
	"fmt"
	"time"

	"ai_interview/internal/domain"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// ResumeRepository stores parsed resumes in PostgreSQL.
type ResumeRepository struct {
	db *gorm.DB
}

// NewResumeRepository creates a ResumeRepository.
func NewResumeRepository(db *gorm.DB) *ResumeRepository {
	return &ResumeRepository{db: db}
}

// GetByHash returns a parsed resume by content hash, or nil when missing.
func (r *ResumeRepository) GetByHash(ctx context.Context, hash string) (*domain.StructuredResume, error) {
	var row ResumeModel
	err := r.db.WithContext(ctx).Where(&ResumeModel{ContentHash: hash}).First(&row).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("[ResumeRepo] get by hash: %w", err)
	}
	resume, err := row.toDomain()
	if err != nil {
		return nil, fmt.Errorf("[ResumeRepo] unmarshal resume: %w", err)
	}
	return resume, nil
}

// GetByUserID returns the latest parsed resume for a user, or nil when missing.
func (r *ResumeRepository) GetByUserID(ctx context.Context, userID string) (*domain.StructuredResume, error) {
	var row ResumeModel
	err := r.db.WithContext(ctx).
		Where(&ResumeModel{UserID: userID}).
		Order(clause.OrderByColumn{Column: clause.Column{Name: "created_at"}, Desc: true}).
		First(&row).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("[ResumeRepo] get by user_id: %w", err)
	}
	resume, err := row.toDomain()
	if err != nil {
		return nil, fmt.Errorf("[ResumeRepo] unmarshal resume: %w", err)
	}
	return resume, nil
}

// Upsert inserts or updates a parsed resume by content_hash.
func (r *ResumeRepository) Upsert(ctx context.Context, userID, hash, s3Key string, resume *domain.StructuredResume) error {
	row, err := newResumeModel(userID, hash, s3Key, resume)
	if err != nil {
		return fmt.Errorf("[ResumeRepo] marshal resume: %w", err)
	}
	err = r.db.WithContext(ctx).Clauses(clause.OnConflict{
		Columns: []clause.Column{{Name: "content_hash"}},
		DoUpdates: clause.Assignments(map[string]any{
			"user_id":    userID,
			"s3_key":     s3Key,
			"updated_at": time.Now(),
		}),
	}).Create(row).Error
	if err != nil {
		return fmt.Errorf("[ResumeRepo] upsert resume: %w", err)
	}
	return nil
}
