package service

import (
	"context"

	"ai_interview/internal/domain"
)

// ReportService 报告模块业务逻辑接口
type ReportService interface {
	// GetStatus 查询报告生成状态：pending / generating / done / failed。
	GetStatus(ctx context.Context, interviewID string) (*ReportStatusResult, error)

	// Get 获取已生成的报告。
	Get(ctx context.Context, interviewID string) (*domain.Report, error)
}

type ReportStatusResult struct {
	Status  string
	Message string
}
