package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"

	"ai_interview/internal/domain"
)

// ErrInterviewNotFound 面试不存在哨兵错误（用于 owner 校验）。
var ErrInterviewNotFound = errors.New("interview not found")

// QuestionnaireRepository 问卷标注 + 关联查询接口，便于 service 层 mock。
type QuestionnaireRepository interface {
	// UpsertBatch 批量插入/更新问卷标注。
	// 冲突键 (interview_id, turn_id)：再次提交同一 turn 走更新，行数不变。
	UpsertBatch(ctx context.Context, items []domain.QuestionnaireResult) error

	// ListByInterview 拉取一次面试的所有问卷标注（用于回显已答情况）。
	ListByInterview(ctx context.Context, interviewID string) ([]domain.QuestionnaireResult, error)

	// InterviewOwner 返回 interview.user_id，不存在返回 ErrInterviewNotFound。
	InterviewOwner(ctx context.Context, interviewID string) (string, error)

	// ExistingTurnIDs 返回给定 interview 下确实存在的 turn_id 子集（用于防伪造校验）。
	ExistingTurnIDs(ctx context.Context, interviewID string, turnIDs []string) ([]string, error)

	// ListTurns 拉取一次面试的所有 turn，按 created_at 升序，前端渲染问卷条目。
	ListTurns(ctx context.Context, interviewID string) ([]domain.InterviewTurn, error)
}

// QuestionnaireRepo 是 QuestionnaireRepository 的 PG 实现。
type QuestionnaireRepo struct {
	db *sql.DB
}

// NewQuestionnaireRepo 创建 QuestionnaireRepo。
func NewQuestionnaireRepo(db *sql.DB) *QuestionnaireRepo {
	return &QuestionnaireRepo{db: db}
}

// UpsertBatch 在单事务内批量 UPSERT，任一失败整批回滚。
func (r *QuestionnaireRepo) UpsertBatch(ctx context.Context, items []domain.QuestionnaireResult) error {
	if len(items) == 0 {
		return nil
	}

	const q = `
		INSERT INTO questionnaire_results
		  (interview_id, turn_id, quality, feedback, user_id, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, NOW(), NOW())
		ON CONFLICT (interview_id, turn_id) DO UPDATE
		  SET quality    = EXCLUDED.quality,
		      feedback   = EXCLUDED.feedback,
		      updated_at = NOW()`

	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("[QuestionnaireRepo] begin tx: %w", err)
	}
	defer func() {
		// 已 Commit 时 Rollback 是 no-op
		_ = tx.Rollback()
	}()

	stmt, err := tx.PrepareContext(ctx, q)
	if err != nil {
		return fmt.Errorf("[QuestionnaireRepo] prepare upsert: %w", err)
	}
	defer stmt.Close()

	for _, it := range items {
		if _, err := stmt.ExecContext(ctx,
			it.InterviewID, it.TurnID, string(it.Quality), it.Feedback, nullableUUID(it.UserID),
		); err != nil {
			return fmt.Errorf("[QuestionnaireRepo] exec upsert (interview=%s turn=%s): %w",
				it.InterviewID, it.TurnID, err)
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("[QuestionnaireRepo] commit: %w", err)
	}
	return nil
}

// ListByInterview 按 interview_id 列出问卷标注。
func (r *QuestionnaireRepo) ListByInterview(ctx context.Context, interviewID string) ([]domain.QuestionnaireResult, error) {
	const q = `
		SELECT id, interview_id, turn_id, quality, COALESCE(feedback, ''), created_at
		FROM questionnaire_results
		WHERE interview_id = $1
		ORDER BY turn_id`

	rows, err := r.db.QueryContext(ctx, q, interviewID)
	if err != nil {
		return nil, fmt.Errorf("[QuestionnaireRepo] query list: %w", err)
	}
	defer rows.Close()

	var out []domain.QuestionnaireResult
	for rows.Next() {
		var it domain.QuestionnaireResult
		var quality string
		if err := rows.Scan(&it.ID, &it.InterviewID, &it.TurnID, &quality, &it.Feedback, &it.CreatedAt); err != nil {
			return nil, fmt.Errorf("[QuestionnaireRepo] scan: %w", err)
		}
		it.Quality = domain.QuestionnaireQuality(quality)
		out = append(out, it)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("[QuestionnaireRepo] rows err: %w", err)
	}
	return out, nil
}

// InterviewOwner 查 interviews.user_id；不存在返回 ErrInterviewNotFound。
func (r *QuestionnaireRepo) InterviewOwner(ctx context.Context, interviewID string) (string, error) {
	const q = `SELECT user_id FROM interviews WHERE id = $1`
	var userID string
	if err := r.db.QueryRowContext(ctx, q, interviewID).Scan(&userID); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return "", ErrInterviewNotFound
		}
		return "", fmt.Errorf("[QuestionnaireRepo] interview owner: %w", err)
	}
	return userID, nil
}

// ExistingTurnIDs 动态展开 IN 占位符以保持驱动无关（不依赖 pq.Array / pgx 数组）。
// 入参 ≤ 50（service 层强制），SQL 注入风险由参数化占位符消除。
func (r *QuestionnaireRepo) ExistingTurnIDs(ctx context.Context, interviewID string, turnIDs []string) ([]string, error) {
	if len(turnIDs) == 0 {
		return nil, nil
	}
	placeholders := make([]string, 0, len(turnIDs))
	args := make([]any, 0, len(turnIDs)+1)
	args = append(args, interviewID)
	for i, id := range turnIDs {
		placeholders = append(placeholders, fmt.Sprintf("$%d", i+2))
		args = append(args, id)
	}
	q := `SELECT turn_id FROM interview_turns WHERE interview_id = $1 AND turn_id IN (` +
		strings.Join(placeholders, ",") + `)`

	rows, err := r.db.QueryContext(ctx, q, args...)
	if err != nil {
		return nil, fmt.Errorf("[QuestionnaireRepo] query existing turns: %w", err)
	}
	defer rows.Close()

	var out []string
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			return nil, fmt.Errorf("[QuestionnaireRepo] scan turn_id: %w", err)
		}
		out = append(out, id)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("[QuestionnaireRepo] rows err: %w", err)
	}
	return out, nil
}

// ListTurns 拉取一次面试的所有 turn。
func (r *QuestionnaireRepo) ListTurns(ctx context.Context, interviewID string) ([]domain.InterviewTurn, error) {
	const q = `
		SELECT id, interview_id, turn_id, stage,
		       COALESCE(question, ''), COALESCE(user_answer, ''), COALESCE(asr_raw, ''),
		       created_at
		FROM interview_turns
		WHERE interview_id = $1
		ORDER BY created_at`

	rows, err := r.db.QueryContext(ctx, q, interviewID)
	if err != nil {
		return nil, fmt.Errorf("[QuestionnaireRepo] query turns: %w", err)
	}
	defer rows.Close()

	var out []domain.InterviewTurn
	for rows.Next() {
		var t domain.InterviewTurn
		if err := rows.Scan(&t.ID, &t.InterviewID, &t.TurnID, &t.Stage,
			&t.Question, &t.UserAnswer, &t.ASRRaw, &t.CreatedAt); err != nil {
			return nil, fmt.Errorf("[QuestionnaireRepo] scan turn: %w", err)
		}
		out = append(out, t)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("[QuestionnaireRepo] rows err: %w", err)
	}
	return out, nil
}

// nullableUUID 空 user_id 写入 NULL，避免 invalid UUID 报错（游客后续可能加 user_id 时复用此分支）。
func nullableUUID(s string) any {
	if strings.TrimSpace(s) == "" {
		return nil
	}
	return s
}
