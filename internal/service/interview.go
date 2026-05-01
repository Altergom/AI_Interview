package service

import (
	"context"

	"ai_interview/internal/domain"
)

// InterviewService 面试模块业务逻辑接口
type InterviewService interface {
	// SetConfig 保存面试岗位和方向配置，存入 Redis。
	SetConfig(ctx context.Context, req InterviewConfigRequest) (configID string, err error)

	// Create 创建面试会话，返回 interview_id 和初始阶段。
	Create(ctx context.Context, userID string) (*InterviewCreateResult, error)

	// ProcessAudio 接收一轮音频数据，触发 ASR → Router → Agent 链路。
	// AI 回复异步通过 SSE 推送，不在此处返回。
	ProcessAudio(ctx context.Context, req AudioRequest) error

	// Finish 结束面试，向 MQ 发布 interview_finished 事件。
	Finish(ctx context.Context, interviewID string) (*InterviewFinishResult, error)

	// GetState 查询面试当前状态。
	GetState(ctx context.Context, interviewID string) (*domain.InterviewState, error)

	// SubmitCode 提交代码，触发 Code Judge Agent → Interview Agent 链路。
	SubmitCode(ctx context.Context, req CodeSubmitRequest) error
}

type InterviewConfigRequest struct {
	UserID    string
	Position  string
	Direction string
}

type InterviewCreateResult struct {
	InterviewID string
	Stage       domain.InterviewStage
	CreatedAt   string
}

type AudioRequest struct {
	InterviewID string
	TurnID      string
	AudioData   []byte
}

type InterviewFinishResult struct {
	InterviewID     string
	FinishedAt      string
	DurationSeconds int64
}

type CodeSubmitRequest struct {
	InterviewID string
	QuestionID  string
	Language    string
	Code        string
}
