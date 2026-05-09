package handler

import (
	"context"
	"net/http"

	"github.com/cloudwego/hertz/pkg/app"

	"ai_interview/internal/domain"
	biz "ai_interview/internal/errors"
	authmw "ai_interview/internal/middleware/auth"
	"ai_interview/internal/service"
)

type resumeHandler struct {
	svc service.ResumeService
}

// PresignUpload GET /v1/resume/upload-url?filename=xxx.pdf
// 生成前端直传 S3 的预签名 PUT URL（5 分钟有效）。
// 响应: { upload_url, object_key }
func (h *resumeHandler) PresignUpload(ctx context.Context, c *app.RequestContext) {
	userID := authmw.GetUserID(c)

	filename := string(c.Query("filename"))
	if filename == "" {
		Fail(ctx, c, http.StatusBadRequest, biz.CodeBadRequest, "filename is required")
		return
	}

	uploadURL, objectKey, err := h.svc.PresignUpload(ctx, userID, filename)
	if err != nil {
		Fail(ctx, c, http.StatusInternalServerError, biz.CodeStorageUploadFailed, "generate upload url failed")
		return
	}

	OK(ctx, c, map[string]string{
		"upload_url": uploadURL,
		"object_key": objectKey,
	})
}

// parseReq POST /v1/resume/parse 请求体。
type parseReq struct {
	// ObjectKey 前端直传 S3 后的对象路径，由 PresignUpload 返回。
	ObjectKey string `json:"object_key"`
}

// Parse POST /v1/resume/parse
// 从 S3 下载 PDF → 文本提取 → SHA-256 去重 → LLM 结构化解析。
// LLM 失败时降级返回空结构体（success=true, data=empty），前端显示空表单。
func (h *resumeHandler) Parse(ctx context.Context, c *app.RequestContext) {
	userID := authmw.GetUserID(c)

	var req parseReq
	if err := c.BindJSON(&req); err != nil || req.ObjectKey == "" {
		Fail(ctx, c, http.StatusBadRequest, biz.CodeBadRequest, "object_key is required")
		return
	}

	resume, err := h.svc.Parse(ctx, userID, req.ObjectKey)
	if err != nil {
		// Parse 只有下载失败才返回 error，其余情况都降级返回空结构体
		HandleErr(ctx, c, err)
		return
	}

	OK(ctx, c, resume)
}

// submitReq POST /v1/resume/submit 请求体（与 domain.StructuredResume 对齐）。
type submitReq struct {
	Skills      []string                  `json:"skills"`
	Projects    []domain.ResumeProject    `json:"projects"`
	Internships []domain.ResumeInternship `json:"internships"`
	Education   domain.ResumeEducation    `json:"education"`
}

// Submit POST /v1/resume/submit
// 保存用户确认（或手填）后的简历，写 PG + Redis。
// 响应: { resume_id }（以 SHA-256 hash 作为幂等 ID）
func (h *resumeHandler) Submit(ctx context.Context, c *app.RequestContext) {
	userID := authmw.GetUserID(c)

	var req submitReq
	if err := c.BindJSON(&req); err != nil {
		Fail(ctx, c, http.StatusBadRequest, biz.CodeBadRequest, "invalid request body")
		return
	}

	resume := domain.StructuredResume{
		UserID:      userID,
		Skills:      req.Skills,
		Projects:    req.Projects,
		Internships: req.Internships,
		Education:   req.Education,
	}

	resumeID, err := h.svc.Submit(ctx, resume)
	if err != nil {
		HandleErr(ctx, c, err)
		return
	}

	OK(ctx, c, map[string]string{"resume_id": resumeID})
}

// Get GET /v1/resume
// 查询当前用户简历（Redis → PG 回填），未找到返回 null data。
func (h *resumeHandler) Get(ctx context.Context, c *app.RequestContext) {
	userID := authmw.GetUserID(c)

	resume, err := h.svc.Get(ctx, userID)
	if err != nil {
		HandleErr(ctx, c, err)
		return
	}
	// resume == nil 说明用户还没有简历，前端据此决定是否跳转上传页
	OK(ctx, c, resume)
}
