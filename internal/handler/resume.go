package handler

import (
	"context"
	"net/http"

	"github.com/cloudwego/hertz/pkg/app"

	"ai_interview/internal/service"
)

type resumeHandler struct {
	svc service.ResumeService
}

// PresignUpload GET /v1/resume/upload-url?filename=xxx.pdf
// 返回前端直传 S3 所需的预签名 PUT URL（5 分钟有效）。
func (h *resumeHandler) PresignUpload(ctx context.Context, c *app.RequestContext) {
	filename := string(c.Query("filename"))
	if filename == "" {
		fail(ctx, c, http.StatusBadRequest, CodeBadRequest, "filename is required")
		return
	}

	// TODO: 从 JWT context 取 userID，当前先用 query 参数占位
	userID := string(c.Query("user_id"))
	if userID == "" {
		fail(ctx, c, http.StatusBadRequest, CodeBadRequest, "user_id is required")
		return
	}

	uploadURL, objectKey, err := h.svc.PresignUpload(ctx, userID, filename)
	if err != nil {
		fail(ctx, c, http.StatusInternalServerError, CodeResumeParseFailed, "generate upload url failed")
		return
	}

	ok(ctx, c, map[string]string{
		"upload_url": uploadURL,
		"object_key": objectKey,
	})
}

// Parse POST /v1/resume/parse  (JSON body: {"object_key": "..."})
// 从 S3 下载 PDF，提取文本，LLM 结构化解析，返回结构化简历。
func (h *resumeHandler) Parse(ctx context.Context, c *app.RequestContext) {
	panic("not implemented")
}

// Submit POST /v1/resume/submit
// 提交用户确认后的简历信息，存入 PG + Redis。
func (h *resumeHandler) Submit(ctx context.Context, c *app.RequestContext) {
	panic("not implemented")
}

// Get GET /v1/resume
// 查询当前用户简历（Redis → PG 回填）。
func (h *resumeHandler) Get(ctx context.Context, c *app.RequestContext) {
	panic("not implemented")
}
