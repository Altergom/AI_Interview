package handler

import (
	"context"

	"github.com/cloudwego/hertz/pkg/app"

	"ai_interview/internal/service"
)

type reportHandler struct {
	svc service.ReportService
}

// Status GET /v1/report/status?interview_id={}
// 查询报告生成状态：pending / generating / done / failed。
func (h *reportHandler) Status(ctx context.Context, c *app.RequestContext) {
	panic("not implemented")
}

// Get GET /v1/report?interview_id={}
// 获取已生成的报告，含各维度评分、总结、优劣势。
func (h *reportHandler) Get(ctx context.Context, c *app.RequestContext) {
	panic("not implemented")
}
