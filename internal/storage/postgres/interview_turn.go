package postgres

import (
	"context"
	"fmt"

	"ai_interview/internal/domain"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// InterviewTurnRepository 面试 turn 落库接口，便于 service 层 mock。
type InterviewTurnRepository interface {
	// SaveTurn 写入一条 turn。同 (interview_id, turn_id) 冲突时 DO NOTHING，
	// 避免重试调用覆盖首次写入内容。
	SaveTurn(ctx context.Context, turn domain.InterviewTurn) error
}

// InterviewTurnRepo InterviewTurnRepository 的 GORM 实现。
type InterviewTurnRepo struct {
	db *gorm.DB
}

// NewInterviewTurnRepo 创建 InterviewTurnRepo。
func NewInterviewTurnRepo(db *gorm.DB) *InterviewTurnRepo {
	return &InterviewTurnRepo{db: db}
}

// SaveTurn 落库一条 turn。同 (interview_id, turn_id) 已存在时 DO NOTHING，
// 保证 SaveTurn 是写入侧的「首次拥有者」语义。
func (r *InterviewTurnRepo) SaveTurn(ctx context.Context, turn domain.InterviewTurn) (err error) {
	defer func() {
		if err != nil {
			err = fmt.Errorf(
				"[InterviewTurnRepo] save turn failed: interview_id=%s turn_id=%s stage=%s: %w",
				turn.InterviewID, turn.TurnID, turn.Stage, err,
			)
		}
	}()

	row := InterviewTurnModel{
		InterviewID: turn.InterviewID,
		TurnID:      turn.TurnID,
		Stage:       turn.Stage,
		Question:    turn.Question,
		UserAnswer:  turn.UserAnswer,
		ASRRaw:      turn.ASRRaw,
	}

	if err = r.db.WithContext(ctx).
		Clauses(clause.OnConflict{DoNothing: true}).
		Create(&row).Error; err != nil {
		return fmt.Errorf("gorm create: %w", err)
	}
	return nil
}
