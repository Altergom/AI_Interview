package service

import (
	"context"
	"errors"

	"ai_interview/internal/domain"
	biz "ai_interview/internal/errors"
	"ai_interview/internal/storage/postgres"
	redisstorage "ai_interview/internal/storage/redis"
)

type reportServiceImpl struct {
	reportRepo *postgres.ReportRepo
	redisCli   *redisstorage.Client
}

func NewReportService(reportRepo *postgres.ReportRepo, redisCli *redisstorage.Client) ReportService {
	return &reportServiceImpl{reportRepo: reportRepo, redisCli: redisCli}
}

func (s *reportServiceImpl) GetStatus(ctx context.Context, interviewID string) (*ReportStatusResult, error) {
	state, err := s.redisCli.GetInterviewState(ctx, interviewID)
	if err == nil && state != nil && state.ReportStatus != "" {
		return &ReportStatusResult{
			Status:  state.ReportStatus,
			Message: reportStatusMessage(state.ReportStatus),
		}, nil
	}

	_, err = s.reportRepo.GetByInterviewID(ctx, interviewID)
	if err == nil {
		return &ReportStatusResult{Status: "done", Message: "报告已生成"}, nil
	}
	if errors.Is(err, postgres.ErrReportNotFound) {
		return &ReportStatusResult{Status: "pending", Message: "报告生成中"}, nil
	}
	return nil, err
}

func (s *reportServiceImpl) Get(ctx context.Context, interviewID string) (*domain.Report, error) {
	report, err := s.reportRepo.GetByInterviewID(ctx, interviewID)
	if errors.Is(err, postgres.ErrReportNotFound) {
		return nil, biz.New(biz.CodeNotFound)
	}
	if err != nil {
		return nil, err
	}
	if report.ErrorMessage != "" {
		return nil, biz.NewMsg(biz.CodeInternal, "报告生成失败: "+report.ErrorMessage)
	}
	return report, nil
}

func reportStatusMessage(status string) string {
	switch status {
	case "pending":
		return "等待生成"
	case "processing":
		return "正在生成"
	case "done":
		return "报告已生成"
	case "failed":
		return "生成失败"
	default:
		return ""
	}
}
