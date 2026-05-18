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

// ErrInterviewNotFound 面试不存在哨兵错误（用于 owner 校验）。
var ErrInterviewNotFound = errors.New("interview not found")

// QuestionnaireRepository 问卷标注 + 关联查询接口，便于 service 层 mock。
type QuestionnaireRepository interface {
	// UpsertBatch 批量插入/更新问卷标注。
	// 冲突键 (interview_id, turn_id)：再次提交同一 turn 走更新，行数不变。
	UpsertBatch(ctx context.Context, items []domain.QuestionnaireResult) error

	// ListByInterview 拉取一次面试的所有问卷标注。
	ListByInterview(ctx context.Context, interviewID string) ([]domain.QuestionnaireResult, error)

	// InterviewOwner 返回 interview.user_id，不存在返回 ErrInterviewNotFound。
	InterviewOwner(ctx context.Context, interviewID string) (string, error)

	// ExistingTurnIDs 返回给定 interview 下确实存在的 turn_id 子集（防伪造）。
	ExistingTurnIDs(ctx context.Context, interviewID string, turnIDs []string) ([]string, error)

	// ListTurns 拉取一次面试的所有 turn，按 created_at 升序。
	ListTurns(ctx context.Context, interviewID string) ([]domain.InterviewTurn, error)
}

// QuestionnaireRepo QuestionnaireRepository 的 GORM 实现。
type QuestionnaireRepo struct {
	db *gorm.DB
}

// NewQuestionnaireRepo 创建 QuestionnaireRepo。
func NewQuestionnaireRepo(db *gorm.DB) *QuestionnaireRepo {
	return &QuestionnaireRepo{db: db}
}

// UpsertBatch 在单事务内批量 UPSERT，任一失败整批回滚。
// ON CONFLICT (interview_id, turn_id) 走 DO UPDATE，保留 created_at，更新 updated_at。
func (r *QuestionnaireRepo) UpsertBatch(ctx context.Context, items []domain.QuestionnaireResult) error {
	if len(items) == 0 {
		return nil
	}

	rows := make([]QuestionnaireResultModel, 0, len(items))
	for _, it := range items {
		rows = append(rows, QuestionnaireResultModel{
			InterviewID: it.InterviewID,
			TurnID:      it.TurnID,
			Quality:     string(it.Quality),
			Feedback:    it.Feedback,
			UserID:      it.UserID,
		})
	}

	err := r.db.WithContext(ctx).Clauses(clause.OnConflict{
		Columns: []clause.Column{{Name: "interview_id"}, {Name: "turn_id"}},
		DoUpdates: clause.Assignments(map[string]any{
			"quality":    gorm.Expr("EXCLUDED.quality"),
			"feedback":   gorm.Expr("EXCLUDED.feedback"),
			"user_id":    gorm.Expr("EXCLUDED.user_id"),
			"updated_at": time.Now(),
		}),
	}).Create(&rows).Error
	if err != nil {
		return fmt.Errorf("[QuestionnaireRepo] upsert batch: %w", err)
	}
	return nil
}

// ListByInterview 按 interview_id 列出问卷标注，按 turn_id 升序。
func (r *QuestionnaireRepo) ListByInterview(ctx context.Context, interviewID string) ([]domain.QuestionnaireResult, error) {
	var rows []QuestionnaireResultModel
	if err := r.db.WithContext(ctx).
		Where(&QuestionnaireResultModel{InterviewID: interviewID}).
		Order("turn_id").
		Find(&rows).Error; err != nil {
		return nil, fmt.Errorf("[QuestionnaireRepo] list by interview: %w", err)
	}

	out := make([]domain.QuestionnaireResult, 0, len(rows))
	for _, row := range rows {
		out = append(out, row.toDomain())
	}
	return out, nil
}

// InterviewOwner 查 interviews.user_id；不存在返回 ErrInterviewNotFound。
func (r *QuestionnaireRepo) InterviewOwner(ctx context.Context, interviewID string) (string, error) {
	var row InterviewModel
	err := r.db.WithContext(ctx).
		Select("user_id").
		Where(&InterviewModel{ID: interviewID}).
		First(&row).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return "", ErrInterviewNotFound
	}
	if err != nil {
		return "", fmt.Errorf("[QuestionnaireRepo] interview owner: %w", err)
	}
	return row.UserID, nil
}

// ExistingTurnIDs 用 GORM IN 子句查询，无需手动拼占位符。
func (r *QuestionnaireRepo) ExistingTurnIDs(ctx context.Context, interviewID string, turnIDs []string) ([]string, error) {
	if len(turnIDs) == 0 {
		return nil, nil
	}

	var out []string
	if err := r.db.WithContext(ctx).
		Model(&InterviewTurnModel{}).
		Where("interview_id = ? AND turn_id IN ?", interviewID, turnIDs).
		Pluck("turn_id", &out).Error; err != nil {
		return nil, fmt.Errorf("[QuestionnaireRepo] existing turn ids: %w", err)
	}
	return out, nil
}

// ListTurns 拉取一次面试的所有 turn。
func (r *QuestionnaireRepo) ListTurns(ctx context.Context, interviewID string) ([]domain.InterviewTurn, error) {
	var rows []InterviewTurnModel
	if err := r.db.WithContext(ctx).
		Where(&InterviewTurnModel{InterviewID: interviewID}).
		Order("created_at").
		Find(&rows).Error; err != nil {
		return nil, fmt.Errorf("[QuestionnaireRepo] list turns: %w", err)
	}

	out := make([]domain.InterviewTurn, 0, len(rows))
	for _, row := range rows {
		out = append(out, row.toDomain())
	}
	return out, nil
}
