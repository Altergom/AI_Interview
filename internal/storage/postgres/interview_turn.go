package postgres

import (
	"context"
	"database/sql"
	"fmt"

	"ai_interview/internal/domain"
)

// InterviewTurnRepository 面试 turn 落库接口，便于 service 层 mock。
type InterviewTurnRepository interface {
	// SaveTurn 写入一条 turn。同 (interview_id, turn_id) 冲突时按需走 UPSERT；
	// 当前业务上 turn_id 单调递增不会重复，DO NOTHING 即可，避免误覆盖历史。
	SaveTurn(ctx context.Context, turn domain.InterviewTurn) error
}

// InterviewTurnRepo InterviewTurnRepository 的 PG 实现。
type InterviewTurnRepo struct {
	db *sql.DB
}

// NewInterviewTurnRepo 创建 InterviewTurnRepo。
func NewInterviewTurnRepo(db *sql.DB) *InterviewTurnRepo {
	return &InterviewTurnRepo{db: db}
}

// SaveTurn 落库一条 turn。
// 设计选择：同 (interview_id, turn_id) 已存在时 DO NOTHING——SaveTurn 是写入侧，
// 重试时同一逻辑 turn 应保持首次写入的内容，避免被后续不完整调用覆盖。
func (r *InterviewTurnRepo) SaveTurn(ctx context.Context, turn domain.InterviewTurn) (err error) {
	defer func() {
		if err != nil {
			err = fmt.Errorf(
				"[InterviewTurnRepo] save turn failed: interview_id=%s turn_id=%s stage=%s: %w",
				turn.InterviewID, turn.TurnID, turn.Stage, err,
			)
		}
	}()

	const q = `
		INSERT INTO interview_turns
		  (interview_id, turn_id, stage, question, user_answer, asr_raw)
		VALUES ($1, $2, $3, $4, $5, $6)
		ON CONFLICT DO NOTHING`

	if _, err = r.db.ExecContext(ctx, q,
		turn.InterviewID, turn.TurnID, turn.Stage,
		nullableText(turn.Question),
		nullableText(turn.UserAnswer),
		nullableText(turn.ASRRaw),
	); err != nil {
		return fmt.Errorf("exec insert: %w", err)
	}
	return nil
}

// nullableText 空字符串写 NULL，保留 TEXT 列允许 NULL 的语义。
func nullableText(s string) any {
	if s == "" {
		return nil
	}
	return s
}
