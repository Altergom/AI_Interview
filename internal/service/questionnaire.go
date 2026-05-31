package service

import (
	"context"

	"ai_interview/internal/domain"
)

// QuestionnaireService 问卷模块业务逻辑接口
type QuestionnaireService interface {
	// Get 获取面试的问卷（每轮问题 + 用户回答）。
	// userID 用于鉴权：仅面试归属者可访问。
	Get(ctx context.Context, userID, interviewID string) ([]domain.InterviewTurn, error)

	// Submit 保存用户对每轮对话的评价。
	Submit(ctx context.Context, req QuestionnaireSubmitRequest) error
}

type QuestionnaireSubmitRequest struct {
	// UserID 提交者，service 层据此做归属校验，由 handler 从 JWT 注入。
	UserID      string
	InterviewID string
	Answers     []QuestionnaireAnswer
}

type QuestionnaireAnswer struct {
	TurnID   string
	Quality  domain.QuestionnaireQuality
	Feedback string
}
