package service

import (
	"context"

	"ai_interview/internal/domain"
)

// ResumeService 简历模块业务逻辑接口
type ResumeService interface {
	// Parse 解析 PDF 简历，同步返回结构化数据。
	Parse(ctx context.Context, fileBytes []byte) (*domain.StructuredResume, error)

	// Submit 保存用户确认后的简历，存入 Redis key: resume:{user_id}，TTL 7天。
	Submit(ctx context.Context, resume domain.StructuredResume) (resumeID string, err error)
}
