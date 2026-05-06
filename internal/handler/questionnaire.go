package handler

import (
	"context"

	"github.com/cloudwego/hertz/pkg/app"

	"ai_interview/internal/service"
)

type questionnaireHandler struct {
	svc service.QuestionnaireService
}

// Get GET /v1/questionnaire?interview_id={}
// 获取面试结束后的问卷列表（每轮问题 + 用户回答）。
func (h *questionnaireHandler) Get(ctx context.Context, c *app.RequestContext) {
	panic("not implemented")
}

// Submit POST /v1/questionnaire/submit
// 提交用户对每轮对话的质量评价（good / bad）及文字反馈。
func (h *questionnaireHandler) Submit(ctx context.Context, c *app.RequestContext) {
	panic("not implemented")
}
