package handler

import (
	"context"

	"github.com/cloudwego/hertz/pkg/app"

	"ai_interview/internal/service"
)

type resumeHandler struct {
	svc service.ResumeService
}

// Parse POST /v1/resume/parse  (multipart/form-data, field: file)
// 同步解析 PDF 简历，返回结构化数据供前端回填表单。
func (h *resumeHandler) Parse(ctx context.Context, c *app.RequestContext) {
	panic("not implemented")
}

// Submit POST /v1/resume/submit
// 提交用户确认后的简历信息，存入 Redis key: resume:{user_id}，TTL 7天。
func (h *resumeHandler) Submit(ctx context.Context, c *app.RequestContext) {
	panic("not implemented")
}
