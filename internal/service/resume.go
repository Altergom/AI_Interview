package service

import (
	"context"

	"ai_interview/internal/domain"
)

// ResumeService 简历模块业务逻辑接口
type ResumeService interface {
	// PresignUpload 生成简历 PDF 直传 S3 的预签名 PUT URL（5 分钟有效）。
	// 返回 uploadURL（前端 PUT 到此地址）和 objectKey（后续 parse 时传回）。
	PresignUpload(ctx context.Context, userID, filename string) (uploadURL, objectKey string, err error)

	// Parse 解析 PDF 简历，同步返回结构化数据。
	// objectKey 为前端直传后 S3 中的路径，由 PresignUpload 返回。
	Parse(ctx context.Context, userID, objectKey string) (*domain.StructuredResume, error)

	// Submit 保存用户确认后的简历，存入 PG 主存储 + Redis 1h 缓存。
	Submit(ctx context.Context, resume domain.StructuredResume) (resumeID string, err error)

	// Get 查询用户当前简历（Redis → PG 回填）。
	Get(ctx context.Context, userID string) (*domain.StructuredResume, error)
}
