package handler

import (
	"context"
	"net/http"

	"github.com/cloudwego/hertz/pkg/app"

	biz "ai_interview/internal/errors"
	"ai_interview/internal/service"
)

type resumeHandler struct {
	svc service.ResumeService
}

// PresignUpload GET /v1/resume/upload-url?filename=xxx.pdf
func (h *resumeHandler) PresignUpload(ctx context.Context, c *app.RequestContext) {
	filename := string(c.Query("filename"))
	if filename == "" {
		Fail(ctx, c, http.StatusBadRequest, biz.CodeBadRequest, "filename is required")
		return
	}

	// TODO: 从 JWT context 取 userID，当前先用 query 参数占位
	userID := string(c.Query("user_id"))
	if userID == "" {
		Fail(ctx, c, http.StatusBadRequest, biz.CodeBadRequest, "user_id is required")
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

// Parse POST /v1/resume/parse
func (h *resumeHandler) Parse(ctx context.Context, c *app.RequestContext) {
	panic("not implemented")
}

// Submit POST /v1/resume/submit
func (h *resumeHandler) Submit(ctx context.Context, c *app.RequestContext) {
	panic("not implemented")
}

// Get GET /v1/resume
func (h *resumeHandler) Get(ctx context.Context, c *app.RequestContext) {
	panic("not implemented")
}
