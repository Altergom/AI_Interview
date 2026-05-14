package postgres

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"

	"ai_interview/internal/domain"
)

// ResumeRepository 简历 PG 存储层。
type ResumeRepository struct {
	db *sql.DB
}

// NewResumeRepository 创建 ResumeRepository。
func NewResumeRepository(db *sql.DB) *ResumeRepository {
	return &ResumeRepository{db: db}
}

// GetByHash 按 content_hash 查询简历，未找到时返回 nil, nil。
// 用于去重：相同文本的 PDF 直接返回已解析结果，不重复调 LLM。
func (r *ResumeRepository) GetByHash(ctx context.Context, hash string) (*domain.StructuredResume, error) {
	const q = `SELECT parsed_data FROM resumes WHERE content_hash = $1 LIMIT 1`

	var raw []byte
	err := r.db.QueryRowContext(ctx, q, hash).Scan(&raw)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("[ResumeRepo] get by hash: %w", err)
	}

	var resume domain.StructuredResume
	if err := json.Unmarshal(raw, &resume); err != nil {
		return nil, fmt.Errorf("[ResumeRepo] unmarshal resume: %w", err)
	}
	return &resume, nil
}

// GetByUserID 查询用户最新一条简历，未找到返回 nil, nil。
func (r *ResumeRepository) GetByUserID(ctx context.Context, userID string) (*domain.StructuredResume, error) {
	const q = `
		SELECT parsed_data FROM resumes
		WHERE user_id = $1
		ORDER BY created_at DESC
		LIMIT 1`

	var raw []byte
	err := r.db.QueryRowContext(ctx, q, userID).Scan(&raw)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("[ResumeRepo] get by user_id: %w", err)
	}

	var resume domain.StructuredResume
	if err := json.Unmarshal(raw, &resume); err != nil {
		return nil, fmt.Errorf("[ResumeRepo] unmarshal resume: %w", err)
	}
	return &resume, nil
}

// Upsert 按 content_hash 插入或更新简历（ON CONFLICT DO UPDATE）。
// s3Key 记录原始 PDF 在 S3 的对象路径（备份追溯用）。
// 同一 hash 的简历若已存在，更新 user_id / s3_key / updated_at（解析结果不变）。
func (r *ResumeRepository) Upsert(ctx context.Context, userID, hash, s3Key string, resume *domain.StructuredResume) error {
	data, err := json.Marshal(resume)
	if err != nil {
		return fmt.Errorf("[ResumeRepo] marshal resume: %w", err)
	}

	const q = `
		INSERT INTO resumes (user_id, content_hash, s3_key, parsed_data)
		VALUES ($1, $2, $3, $4)
		ON CONFLICT (content_hash) DO UPDATE
		  SET user_id    = EXCLUDED.user_id,
		      s3_key     = EXCLUDED.s3_key,
		      updated_at = NOW()`

	if _, err := r.db.ExecContext(ctx, q, userID, hash, s3Key, data); err != nil {
		return fmt.Errorf("[ResumeRepo] upsert resume: %w", err)
	}
	return nil
}
