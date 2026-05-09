package service

import (
	"context"
	"crypto/sha256"
	"fmt"
	"time"

	"ai_interview/internal/config"
	"ai_interview/internal/domain"
	"ai_interview/internal/einocore"
	"ai_interview/internal/llm"
	"ai_interview/internal/log"
	resumepdf "ai_interview/internal/resume/pdf"
	pgstore "ai_interview/internal/storage/postgres"
	redistore "ai_interview/internal/storage/redis"
	s3store "ai_interview/internal/storage/s3"
)

const (
	resumePresignTTL    = 5 * time.Minute // 预签名 URL 有效期
	resumeRedisTTL      = 1 * time.Hour   // Redis 缓存 TTL
	resumeMaxConcurrent = 5               // 最大并发解析数（PDF 提取 + LLM 均为 CPU/IO 密集）
)

// resumeParseSystemPrompt 简历解析系统提示词。
const resumeParseSystemPrompt = `你是一名专业的简历信息提取助手。
请从用户提供的简历文本中提取结构化信息，严格按照以下 JSON 格式输出，不要包含任何多余文字或 markdown 标记：

{
  "user_id": "",
  "skills": ["技能1", "技能2"],
  "projects": [
    {
      "name": "项目名称",
      "tech_stack": ["技术1", "技术2"],
      "description": "项目描述",
      "highlights": ["亮点1", "亮点2"]
    }
  ],
  "internships": [
    {
      "company": "公司名称",
      "role": "岗位名称",
      "description": "工作描述"
    }
  ],
  "education": {
    "school": "学校名称",
    "major": "专业名称",
    "graduation": "毕业年份，如 2024-06"
  }
}

注意：
- user_id 保持为空字符串，由系统填充
- 若某字段在简历中找不到对应信息，使用空字符串或空数组
- skills 只保留技术技能（编程语言、框架、工具等），不包含软技能
- 只输出 JSON，不要解释`

// resumeService 实现 ResumeService。
type resumeService struct {
	s3  *s3store.Client
	db  *pgstore.ResumeRepository
	rdb *redistore.Client
	cfg *config.Config
	// sem 信号量：限制最多 resumeMaxConcurrent 个并发解析（PDF 提取 + LLM）
	sem chan struct{}
}

// NewResumeService 构造 ResumeService 实例。
func NewResumeService(
	s3Client *s3store.Client,
	repo *pgstore.ResumeRepository,
	rdb *redistore.Client,
	cfg *config.Config,
) ResumeService {
	return &resumeService{
		s3:  s3Client,
		db:  repo,
		rdb: rdb,
		cfg: cfg,
		sem: make(chan struct{}, resumeMaxConcurrent),
	}
}

// PresignUpload 生成简历 PDF 直传 S3 的预签名 PUT URL（5 分钟有效）。
// 前端拿到 URL 后直接 PUT 文件，PDF 原始文件保存在 S3 的 objectKey 路径下。
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

// Parse 从 S3 下载 PDF → 提取文本 → SHA-256 去重 → LLM 结构化解析。
//
// 并发控制：信号量限制最多 resumeMaxConcurrent(5) 个并发解析，
// 超过则阻塞等待，不返回错误（由调用超时控制）。
//
// 大小限制：PDF 提取阶段内置 3MB 硬限（io.LimitReader）。
//
// PDF 备份：原始文件通过预签名 URL 由前端直传 S3，objectKey 落库到 resumes.s3_key，
// 后续可通过 PresignGetURL(objectKey) 追溯原始文件。
//
// 去重路径：content_hash 命中 PG → 直接返回已解析结果，跳过 LLM 调用。
// 降级路径：LLM 多次失败 → 返回空结构体，不阻塞前端。
func (s *resumeService) Parse(ctx context.Context, userID, objectKey string) (*domain.StructuredResume, error) {
	// 获取信号量（并发限制）
	select {
	case s.sem <- struct{}{}:
		defer func() { <-s.sem }()
	case <-ctx.Done():
		return nil, ctx.Err()
	}

	// 步骤 1：从 S3 下载 PDF（原始文件已由前端直传，这里是读取）
	rc, err := s.s3.Download(ctx, objectKey)
	if err != nil {
		log.Errorf("[ResumeService] download pdf key=%s: %v", objectKey, err)
		return nil, fmt.Errorf("download pdf: %w", err)
	}
	defer rc.Close()

	// 步骤 2：逐页提取文本（内置 3MB 大小限制，超出返回 ErrFileTooLarge）
	text, err := resumepdf.ExtractText(rc)
	if err != nil {
		log.Errorf("[ResumeService] extract text key=%s: %v", objectKey, err)
		return emptyResume(userID), nil // PDF 损坏/超限，降级
	}
	log.Infof("[ResumeService] pdf text extracted key=%s, chars=%d", objectKey, len(text))

	// 步骤 3：SHA-256 去重（命中则跳过 LLM，直接返回）
	hash := sha256Hex(text)
	if cached, err := s.db.GetByHash(ctx, hash); err == nil && cached != nil {
		cached.UserID = userID
		// 更新 PG 中的 user_id 绑定和 s3_key（新用户上传相同简历）
		_ = s.db.Upsert(ctx, userID, hash, objectKey, cached)
		_ = s.rdb.SaveResume(ctx, cached, s.resumeTTL())
		log.Infof("[ResumeService] resume cache hit hash=%.8s user=%s", hash, userID)
		return cached, nil
	}

	// 步骤 4：LLM 结构化解析（经 StructuredOutputInvoker，最多重试 3 次）
	model, err := llm.Registry.NewChatModel(ctx, llm.RoleEvaluator)
	if err != nil {
		log.Errorf("[ResumeService] new chat model: %v", err)
		return emptyResume(userID), nil
	}

	invoker := einocore.NewStructuredOutputInvoker(model, 3)
	var result domain.StructuredResume
	if err := invoker.Invoke(ctx, resumeParseSystemPrompt, text, &result); err != nil {
		log.Warnf("[ResumeService] LLM parse failed key=%s (fallback to empty): %v", objectKey, err)
		return emptyResume(userID), nil
	}
	result.UserID = userID

	// 步骤 5：写入 PG（去重 upsert，记录 s3_key 备份路径）
	if err := s.db.Upsert(ctx, userID, hash, objectKey, &result); err != nil {
		log.Errorf("[ResumeService] upsert resume to PG: %v", err)
		// 写 PG 失败不阻断
	}

	log.Infof("[ResumeService] resume parsed successfully user=%s hash=%.8s", userID, hash)
	return &result, nil
}

// Submit 保存用户确认后的简历，写 PG 主存储 + Redis 1h 缓存。
func (s *resumeService) Submit(ctx context.Context, resume domain.StructuredResume) (resumeID string, err error) {
	if resume.UserID == "" {
		return "", fmt.Errorf("user_id is required")
	}

	serialized := fmt.Sprintf("%v", resume)
	hash := sha256Hex(serialized)

	if err := s.db.Upsert(ctx, resume.UserID, hash, "", &resume); err != nil {
		return "", fmt.Errorf("save resume to PG: %w", err)
	}

	if err := s.rdb.SaveResume(ctx, &resume, s.resumeTTL()); err != nil {
		log.Warnf("[ResumeService] save resume to Redis: %v", err)
	}

	log.Infof("[ResumeService] resume submitted user=%s", resume.UserID)
	return hash, nil
}

// Get 查询用户当前简历，优先读 Redis，未命中则回源 PG 并回填缓存。
func (s *resumeService) Get(ctx context.Context, userID string) (*domain.StructuredResume, error) {
	if cached, err := s.rdb.GetResume(ctx, userID); err == nil && cached != nil {
		return cached, nil
	}

	resume, err := s.db.GetByUserID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("get resume from PG: %w", err)
	}
	if resume == nil {
		return nil, nil
	}

	if err := s.rdb.SaveResume(ctx, resume, s.resumeTTL()); err != nil {
		log.Warnf("[ResumeService] refill redis resume user=%s: %v", userID, err)
	}
	return resume, nil
}

// resumeTTL 返回简历 Redis 缓存 TTL。
func (s *resumeService) resumeTTL() time.Duration {
	if s.cfg != nil && s.cfg.ResumeRedisTTL > 0 {
		return s.cfg.ResumeRedisTTL
	}
	return resumeRedisTTL
}

// emptyResume 返回只有 UserID 的空简历，用于 LLM 失败降级。
func emptyResume(userID string) *domain.StructuredResume {
	return &domain.StructuredResume{
		UserID:      userID,
		Skills:      []string{},
		Projects:    []domain.ResumeProject{},
		Internships: []domain.ResumeInternship{},
		Education:   domain.ResumeEducation{},
	}
}

// sha256Hex 计算字符串的 SHA-256 十六进制摘要（64 字符）。
func sha256Hex(s string) string {
	h := sha256.Sum256([]byte(s))
	return fmt.Sprintf("%x", h)
}
