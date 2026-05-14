package postgres

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"

	"ai_interview/internal/domain"
	"ai_interview/internal/log"
)

// BankQuestionRepo 题库 PG 存储层。
type BankQuestionRepo struct {
	db *sql.DB
}

// NewBankQuestionRepo 创建 BankQuestionRepo。
func NewBankQuestionRepo(db *sql.DB) *BankQuestionRepo {
	return &BankQuestionRepo{db: db}
}

// Insert 新增一条题目，返回生成的 UUID。
// vec_status 默认由数据库 DEFAULT 'pending' 设置。
func (r *BankQuestionRepo) Insert(ctx context.Context, q *domain.BankQuestionRecord) (string, error) {
	tagsJSON, err := json.Marshal(q.Tags)
	if err != nil {
		return "", fmt.Errorf("[BankQuestionRepo] marshal tags: %w", err)
	}
	conceptsJSON, err := json.Marshal(q.RelatedConcepts)
	if err != nil {
		return "", fmt.Errorf("[BankQuestionRepo] marshal related_concepts: %w", err)
	}
	followupJSON, err := json.Marshal(q.FollowupQuestionIDs)
	if err != nil {
		return "", fmt.Errorf("[BankQuestionRepo] marshal followup_question_ids: %w", err)
	}

	const stmt = `
		INSERT INTO bank_questions
			(question, standard_answer, tags, related_concepts, followup_question_ids, difficulty)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id`

	var id string
	err = r.db.QueryRowContext(ctx, stmt,
		q.Question,
		q.StandardAnswer,
		tagsJSON,
		conceptsJSON,
		followupJSON,
		string(q.Difficulty),
	).Scan(&id)
	if err != nil {
		return "", fmt.Errorf("[BankQuestionRepo] insert: %w", err)
	}
	log.Debugf("[BankQuestionRepo] inserted question id=%s", id)
	return id, nil
}

// GetByID 按主键查询，未找到返回 nil, nil。
func (r *BankQuestionRepo) GetByID(ctx context.Context, id string) (*domain.BankQuestionRecord, error) {
	const q = `
		SELECT id, question, standard_answer, tags, related_concepts,
		       followup_question_ids, difficulty, vec_status, created_at, updated_at
		FROM bank_questions
		WHERE id = $1`

	row := r.db.QueryRowContext(ctx, q, id)
	rec, err := scanBankQuestion(row)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("[BankQuestionRepo] get by id: %w", err)
	}
	return rec, nil
}

// UpdateVecStatus 更新向量化状态。
func (r *BankQuestionRepo) UpdateVecStatus(ctx context.Context, id string, status domain.VecStatus) error {
	const stmt = `UPDATE bank_questions SET vec_status = $1, updated_at = NOW() WHERE id = $2`
	res, err := r.db.ExecContext(ctx, stmt, string(status), id)
	if err != nil {
		return fmt.Errorf("[BankQuestionRepo] update vec_status id=%s: %w", id, err)
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return fmt.Errorf("[BankQuestionRepo] question not found: %s", id)
	}
	return nil
}

// ListPending 批量拉取待向量化的题目，按 created_at 升序，limit 条。
func (r *BankQuestionRepo) ListPending(ctx context.Context, limit int) ([]*domain.BankQuestionRecord, error) {
	const q = `
		SELECT id, question, standard_answer, tags, related_concepts,
		       followup_question_ids, difficulty, vec_status, created_at, updated_at
		FROM bank_questions
		WHERE vec_status = 'pending'
		ORDER BY created_at ASC
		LIMIT $1`

	rows, err := r.db.QueryContext(ctx, q, limit)
	if err != nil {
		return nil, fmt.Errorf("[BankQuestionRepo] list pending: %w", err)
	}
	defer rows.Close()

	var records []*domain.BankQuestionRecord
	for rows.Next() {
		rec, err := scanBankQuestionRow(rows)
		if err != nil {
			return nil, fmt.Errorf("[BankQuestionRepo] scan pending row: %w", err)
		}
		records = append(records, rec)
	}
	return records, rows.Err()
}

// ListByTags 按标签 GIN 包含查询，最多返回 limit 条（vec_status=done）。
func (r *BankQuestionRepo) ListByTags(ctx context.Context, tags []string, limit int) ([]*domain.BankQuestionRecord, error) {
	tagsJSON, err := json.Marshal(tags)
	if err != nil {
		return nil, fmt.Errorf("[BankQuestionRepo] marshal tags filter: %w", err)
	}

	const q = `
		SELECT id, question, standard_answer, tags, related_concepts,
		       followup_question_ids, difficulty, vec_status, created_at, updated_at
		FROM bank_questions
		WHERE tags @> $1::jsonb
		  AND vec_status = 'done'
		ORDER BY created_at DESC
		LIMIT $2`

	rows, err := r.db.QueryContext(ctx, q, string(tagsJSON), limit)
	if err != nil {
		return nil, fmt.Errorf("[BankQuestionRepo] list by tags: %w", err)
	}
	defer rows.Close()

	var records []*domain.BankQuestionRecord
	for rows.Next() {
		rec, err := scanBankQuestionRow(rows)
		if err != nil {
			return nil, fmt.Errorf("[BankQuestionRepo] scan tag row: %w", err)
		}
		records = append(records, rec)
	}
	return records, rows.Err()
}

// ------- 内部扫描辅助 -------

type rowScanner interface {
	Scan(dest ...any) error
}

func scanBankQuestion(row *sql.Row) (*domain.BankQuestionRecord, error) {
	var (
		rec         domain.BankQuestionRecord
		tagsRaw     []byte
		conceptsRaw []byte
		followupRaw []byte
		difficulty  string
		vecStatus   string
	)
	if err := row.Scan(
		&rec.ID, &rec.Question, &rec.StandardAnswer,
		&tagsRaw, &conceptsRaw, &followupRaw,
		&difficulty, &vecStatus,
		&rec.CreatedAt, &rec.UpdatedAt,
	); err != nil {
		return nil, err
	}
	return parseBankQuestionFields(&rec, tagsRaw, conceptsRaw, followupRaw, difficulty, vecStatus)
}

func scanBankQuestionRow(rows *sql.Rows) (*domain.BankQuestionRecord, error) {
	var (
		rec         domain.BankQuestionRecord
		tagsRaw     []byte
		conceptsRaw []byte
		followupRaw []byte
		difficulty  string
		vecStatus   string
	)
	if err := rows.Scan(
		&rec.ID, &rec.Question, &rec.StandardAnswer,
		&tagsRaw, &conceptsRaw, &followupRaw,
		&difficulty, &vecStatus,
		&rec.CreatedAt, &rec.UpdatedAt,
	); err != nil {
		return nil, err
	}
	return parseBankQuestionFields(&rec, tagsRaw, conceptsRaw, followupRaw, difficulty, vecStatus)
}

func parseBankQuestionFields(rec *domain.BankQuestionRecord, tagsRaw, conceptsRaw, followupRaw []byte, difficulty, vecStatus string) (*domain.BankQuestionRecord, error) {
	if err := json.Unmarshal(tagsRaw, &rec.Tags); err != nil {
		return nil, fmt.Errorf("unmarshal tags: %w", err)
	}
	if err := json.Unmarshal(conceptsRaw, &rec.RelatedConcepts); err != nil {
		return nil, fmt.Errorf("unmarshal related_concepts: %w", err)
	}
	if err := json.Unmarshal(followupRaw, &rec.FollowupQuestionIDs); err != nil {
		return nil, fmt.Errorf("unmarshal followup_question_ids: %w", err)
	}
	rec.Difficulty = domain.Difficulty(difficulty)
	rec.VecStatus = domain.VecStatus(vecStatus)
	return rec, nil
}
