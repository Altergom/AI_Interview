package service

import (
	"context"
	"fmt"
	"time"

	"ai_interview/internal/domain"
	"ai_interview/internal/einocore"
	"ai_interview/internal/llm"
	"ai_interview/internal/log"
	resumepdf "ai_interview/internal/resume/pdf"
	s3store "ai_interview/internal/storage/s3"
)

const (
	resumePresignTTL = 5 * time.Minute // 预签名 URL 有效期，固定 5 分钟
)

// resumeParseSystemPrompt 简历解析系统提示词。
// 要求 LLM 严格输出 JSON，字段与 domain.StructuredResume 对应。
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
// 降级策略：LLM 解析失败时返回 UserID 已填充的空 StructuredResume，
// 不返回 error，前端收到空结构体后显示空表单让用户手填。
func (s *resumeService) Parse(ctx context.Context, userID, objectKey string) (*domain.StructuredResume, error) {
	// 步骤 1：从 S3 下载 PDF
	rc, err := s.s3.Download(ctx, objectKey)
	if err != nil {
		log.Errorf("[ResumeService] download pdf key=%s: %v", objectKey, err)
		return nil, fmt.Errorf("download pdf: %w", err)
	}
	defer rc.Close()

	// 步骤 2：逐页提取文本（内置 3MB 大小限制）
	text, err := resumepdf.ExtractText(rc)
	if err != nil {
		log.Errorf("[ResumeService] extract text key=%s: %v", objectKey, err)
		// PDF 本身损坏或超限，直接降级返回空简历
		return emptyResume(userID), nil
	}
	log.Infof("[ResumeService] pdf text extracted key=%s, chars=%d", objectKey, len(text))

	// 步骤 3：LLM 结构化解析（经 StructuredOutputInvoker，最多重试 3 次）
	model, err := llm.Registry.NewChatModel(ctx, llm.RoleEvaluator)
	if err != nil {
		log.Errorf("[ResumeService] new chat model: %v", err)
		return emptyResume(userID), nil
	}

	invoker := einocore.NewStructuredOutputInvoker(model, 3)
	var result domain.StructuredResume
	if err := invoker.Invoke(ctx, resumeParseSystemPrompt, text, &result); err != nil {
		// 降级：LLM 多次失败，返回空结构体，前端显示空表单
		log.Warnf("[ResumeService] LLM parse failed for key=%s (fallback to empty): %v", objectKey, err)
		return emptyResume(userID), nil
	}

	// 补填 userID（prompt 要求 LLM 保持 user_id 为空）
	result.UserID = userID
	log.Infof("[ResumeService] resume parsed successfully for user %s", userID)
	return &result, nil
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

// Submit 保存用户确认后的简历。
// TODO(task-next): 实现 PG 主存储 + Redis 缓存。
func (s *resumeService) Submit(ctx context.Context, resume domain.StructuredResume) (resumeID string, err error) {
	return "", fmt.Errorf("not implemented")
}

// Get 查询用户当前简历（Redis → PG 回填）。
// TODO(task-next): 实现缓存 + DB 查询。
func (s *resumeService) Get(ctx context.Context, userID string) (*domain.StructuredResume, error) {
	return nil, fmt.Errorf("not implemented")
}
