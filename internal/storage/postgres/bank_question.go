package postgres

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"ai_interview/internal/domain"
	"ai_interview/internal/log"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// BankQuestionRepo stores bank question metadata in PostgreSQL.
type BankQuestionRepo struct {
	db *gorm.DB
}

// NewBankQuestionRepo creates a BankQuestionRepo.
func NewBankQuestionRepo(db *gorm.DB) *BankQuestionRepo {
	return &BankQuestionRepo{db: db}
}

// Insert inserts one question and returns the database-generated UUID.
func (r *BankQuestionRepo) Insert(ctx context.Context, q *domain.BankQuestionRecord) (string, error) {
	row, err := newBankQuestionModel(q)
	if err != nil {
		return "", fmt.Errorf("[BankQuestionRepo] build model: %w", err)
	}
	err = r.db.WithContext(ctx).
		Clauses(clause.Returning{Columns: []clause.Column{{Name: "id"}}}).
		Create(row).Error
	if err != nil {
		return "", fmt.Errorf("[BankQuestionRepo] insert: %w", err)
	}
	log.Debugf("[BankQuestionRepo] inserted question id=%s", row.ID)
	return row.ID, nil
}

// GetByID returns a question by primary key, or nil when missing.
func (r *BankQuestionRepo) GetByID(ctx context.Context, id string) (*domain.BankQuestionRecord, error) {
	var row BankQuestionModel
	err := r.db.WithContext(ctx).Where(&BankQuestionModel{ID: id}).First(&row).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("[BankQuestionRepo] get by id: %w", err)
	}
	rec, err := row.toDomain()
	if err != nil {
		return nil, fmt.Errorf("[BankQuestionRepo] parse by id: %w", err)
	}
	return rec, nil
}

// UpdateVecStatus updates the async vectorization status.
func (r *BankQuestionRepo) UpdateVecStatus(ctx context.Context, id string, status domain.VecStatus) error {
	res := r.db.WithContext(ctx).
		Model(&BankQuestionModel{}).
		Where(&BankQuestionModel{ID: id}).
		Updates(map[string]any{
			"vec_status": string(status),
			"updated_at": time.Now(),
		})
	if res.Error != nil {
		return fmt.Errorf("[BankQuestionRepo] update vec_status id=%s: %w", id, res.Error)
	}
	if res.RowsAffected == 0 {
		return fmt.Errorf("[BankQuestionRepo] question not found: %s", id)
	}
	return nil
}

// ListPending lists pending vectorization tasks by creation time.
func (r *BankQuestionRepo) ListPending(ctx context.Context, limit int) ([]*domain.BankQuestionRecord, error) {
	var rows []BankQuestionModel
	err := r.db.WithContext(ctx).
		Where(&BankQuestionModel{VecStatus: string(domain.VecStatusPending)}).
		Order(clause.OrderByColumn{Column: clause.Column{Name: "created_at"}}).
		Limit(limit).
		Find(&rows).Error
	if err != nil {
		return nil, fmt.Errorf("[BankQuestionRepo] list pending: %w", err)
	}
	return bankQuestionModelsToDomain(rows)
}

// ListByTags uses the chosen JSONB contains query with GORM parameter binding.
func (r *BankQuestionRepo) ListByTags(ctx context.Context, tags []string, limit int) ([]*domain.BankQuestionRecord, error) {
	tagsJSON, err := json.Marshal(tags)
	if err != nil {
		return nil, fmt.Errorf("[BankQuestionRepo] marshal tags filter: %w", err)
	}

	var rows []BankQuestionModel
	err = r.db.WithContext(ctx).
		Where("tags @> ?::jsonb", string(tagsJSON)).
		Where(&BankQuestionModel{VecStatus: string(domain.VecStatusDone)}).
		Order(clause.OrderByColumn{Column: clause.Column{Name: "created_at"}, Desc: true}).
		Limit(limit).
		Find(&rows).Error
	if err != nil {
		return nil, fmt.Errorf("[BankQuestionRepo] list by tags: %w", err)
	}
	return bankQuestionModelsToDomain(rows)
}

// ExistsByQuestion 检查已存在相同题目的记录，用于跨运行去重。
func (r *BankQuestionRepo) ExistsByQuestion(ctx context.Context, question string) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Model(&BankQuestionModel{}).
		Where("question = ?", question).
		Count(&count).Error
	if err != nil {
		return false, fmt.Errorf("[BankQuestionRepo] exists by question: %w", err)
	}
	return count > 0, nil
}
