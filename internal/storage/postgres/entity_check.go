package postgres

import (
	"context"
	"fmt"

	"gorm.io/gorm"
)

// EntityChecker validates that async worker targets still exist.
type EntityChecker struct {
	db *gorm.DB
}

// NewEntityChecker creates an EntityChecker.
func NewEntityChecker(db *gorm.DB) *EntityChecker {
	return &EntityChecker{db: db}
}

// InterviewExists checks whether an interview exists.
func (c *EntityChecker) InterviewExists(ctx context.Context, interviewID string) (bool, error) {
	return c.exists(ctx, &InterviewModel{}, interviewID, "interviews")
}

// BankQuestionExists checks whether a bank question exists.
func (c *EntityChecker) BankQuestionExists(ctx context.Context, questionID string) (bool, error) {
	return c.exists(ctx, &BankQuestionModel{}, questionID, "bank_questions")
}

func (c *EntityChecker) exists(ctx context.Context, model any, id, name string) (bool, error) {
	var count int64
	err := c.db.WithContext(ctx).Model(model).Where(map[string]any{"id": id}).Count(&count).Error
	if err != nil {
		return false, fmt.Errorf("check %s exists id=%s: %w", name, id, err)
	}
	return count > 0, nil
}
