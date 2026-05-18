package handler

import (
	"context"
	"net/http"

	"github.com/cloudwego/hertz/pkg/app"

	biz "ai_interview/internal/errors"
	"ai_interview/internal/service"
)

type reportHandler struct {
	svc service.ReportService
}

// Status GET /v1/report/status?interview_id={}
func (h *reportHandler) Status(ctx context.Context, c *app.RequestContext) {
	interviewID := string(c.Query("interview_id"))
	if interviewID == "" {
		Fail(ctx, c, http.StatusBadRequest, biz.CodeBadRequest, "interview_id is required")
		return
	}

	result, err := h.svc.GetStatus(ctx, interviewID)
	if err != nil {
		HandleErr(ctx, c, err)
		return
	}
	OK(ctx, c, result)
}

// Get GET /v1/report?interview_id={}
func (h *reportHandler) Get(ctx context.Context, c *app.RequestContext) {
	interviewID := string(c.Query("interview_id"))
	if interviewID == "" {
		Fail(ctx, c, http.StatusBadRequest, biz.CodeBadRequest, "interview_id is required")
		return
	}

	report, err := h.svc.Get(ctx, interviewID)
	if err != nil {
		HandleErr(ctx, c, err)
		return
	}
	OK(ctx, c, report)
}
