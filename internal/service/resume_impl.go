package service

import (
	"context"
	"fmt"
	"time"

	"ai_interview/internal/domain"
	"ai_interview/internal/log"
	resumepdf "ai_interview/internal/resume/pdf"
	s3store "ai_interview/internal/storage/s3"
)

const (
	resumePresignTTL = 5 * time.Minute // 预签名 URL 有效期，固定 5 分钟
)

// resumeService 实现 ResumeService。
// 依赖由 NewResumeService 注入，所有字段不可为 nil。
type resumeService struct {
	s3 *s3store.Client
	// db 和 rdb 在后续任务中注入，此处预留字段
	// db  *postgres.DB
	// rdb *redis.Client
}

// NewResumeService 构造 ResumeService 实例。
func NewResumeService(s3Client *s3store.Client) ResumeService {
	return &resumeService{
		s3: s3Client,
	}
}

// PresignUpload 生成简历 PDF 直传 S3 的预签名 PUT URL（5 分钟有效）。
//
// 流程：
//  1. 根据 userID + filename 构造 S3 object key
//  2. 调用 S3 client 生成预签名 PUT URL
//  3. 返回 URL 和 key 给前端；前端直接 PUT 文件，不经过服务端
//
// 安全：URL 仅限 PUT application/pdf，5 分钟过期，避免长时间暴露。
func (s *resumeService) PresignUpload(ctx context.Context, userID, filename string) (uploadURL, objectKey string, err error) {
	key := s3store.ResumeObjectKey(userID, filename)

	url, err := s.s3.PresignPutURL(ctx, key, "application/pdf", resumePresignTTL)
	if err != nil {
		log.Errorf("[ResumeService] presign upload for user %s: %v", userID, err)
		return "", "", fmt.Errorf("presign upload: %w", err)
	}

	log.Infof("[ResumeService] presign upload url generated for user %s, key=%s", userID, key)
	return url, key, nil
}

// Parse 从 S3 下载 PDF → 逐页提取文本 → LLM 结构化解析。
//
// 当前实现：步骤 1（下载）+ 步骤 2（PDF 文本提取）已完成。
// TODO(task-3): 步骤 3 LLM 结构化解析（经 StructuredOutputInvoker）。
func (s *resumeService) Parse(ctx context.Context, userID, objectKey string) (*domain.StructuredResume, error) {
	// 步骤 1：从 S3 下载 PDF
	rc, err := s.s3.Download(ctx, objectKey)
	if err != nil {
		log.Errorf("[ResumeService] download pdf from S3 key=%s: %v", objectKey, err)
		return nil, fmt.Errorf("download pdf: %w", err)
	}
	defer rc.Close()

	// 步骤 2：逐页提取文本（内置 3MB 大小限制，超出返回 ErrFileTooLarge）
	text, err := resumepdf.ExtractText(rc)
	if err != nil {
		log.Errorf("[ResumeService] extract text from pdf key=%s: %v", objectKey, err)
		return nil, fmt.Errorf("extract pdf text: %w", err)
	}

	log.Infof("[ResumeService] pdf text extracted, key=%s, length=%d chars", objectKey, len(text))

	// 步骤 3：LLM 结构化解析（下一个任务实现）
	// 当前返回包含原始文本的占位结果，确保后续任务可以接入
	_ = text // 文本已提取，等待 LLM 解析接入
	return nil, fmt.Errorf("LLM parsing not yet implemented (text extracted: %d chars)", len(text))
}

// Submit 保存用户确认后的简历。
// TODO(task-5): 实现 PG 主存储 + Redis 缓存。
func (s *resumeService) Submit(ctx context.Context, resume domain.StructuredResume) (resumeID string, err error) {
	return "", fmt.Errorf("not implemented")
}

// Get 查询用户当前简历（Redis → PG 回填）。
// TODO(task-5): 实现缓存 + DB 查询。
func (s *resumeService) Get(ctx context.Context, userID string) (*domain.StructuredResume, error) {
	return nil, fmt.Errorf("not implemented")
}
