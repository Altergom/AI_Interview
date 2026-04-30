package service

import (
	"context"

	"ai_interview/internal/domain"
)

// QuestionnaireService 问卷模块业务逻辑接口
type QuestionnaireService interface {
	// Get 获取面试的问卷（每轮问题 + 用户回答）。
	Get(ctx context.Context, interviewID string) ([]domain.InterviewTurn, error)

	// Submit 保存用户对每轮对话的评价。
	Submit(ctx context.Context, req QuestionnaireSubmitRequest) error
}

type QuestionnaireSubmitRequest struct {
	InterviewID string
	Answers     []QuestionnaireAnswer
}

type QuestionnaireAnswer struct {
	TurnID   string
	Quality  domain.QuestionnaireQuality
	Feedback string
}
