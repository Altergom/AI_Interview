package service

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"

	"ai_interview/internal/domain"
	"ai_interview/internal/einocore/compose"
	"ai_interview/internal/log"
	redisstorage "ai_interview/internal/storage/redis"
)

// interviewServiceImpl InterviewService 的实现
type interviewServiceImpl struct {
	sessionManager *SessionManager
	graph          *compose.InterviewGraph
	redisCli       *redisstorage.Client
	stateTTL       time.Duration
}

// NewInterviewService 创建 InterviewService 实例，所有依赖从外部注入。
func NewInterviewService(
	sessionManager *SessionManager,
	graph *compose.InterviewGraph,
	redisCli *redisstorage.Client,
	stateTTL time.Duration,
) InterviewService {
	return &interviewServiceImpl{
		sessionManager: sessionManager,
		graph:          graph,
		redisCli:       redisCli,
		stateTTL:       stateTTL,
	}
}

// SetConfig 保存面试岗位和方向配置到 Redis，返回 interview_id。
// 前端拿到 interview_id 后再调 Create 创建 session。
func (s *interviewServiceImpl) SetConfig(ctx context.Context, req InterviewConfigRequest) (string, error) {
	if req.Direction == "" {
		return "", fmt.Errorf("[interview] direction is required")
	}

	interviewID := uuid.New().String()

	cfg := redisstorage.InterviewConfig{
		Direction: req.Direction,
		Position:  req.Position,
	}
	if err := s.redisCli.SaveInterviewConfig(ctx, interviewID, cfg, s.stateTTL); err != nil {
		return "", fmt.Errorf("[interview] save config: %w", err)
	}

	return interviewID, nil
}

// Create 创建面试 session，interviewID 由 SetConfig 返回，此处复用。
func (s *interviewServiceImpl) Create(ctx context.Context, interviewID, userID string) (*InterviewCreateResult, error) {
	// 校验 config 存在
	cfg, err := s.redisCli.GetInterviewConfig(ctx, interviewID)
	if err != nil {
		return nil, fmt.Errorf("[interview] get config: %w", err)
	}
	if cfg == nil {
		return nil, fmt.Errorf("[interview] config not found for interview_id=%s, call SetConfig first", interviewID)
	}

	// 创建 session
	if err := s.sessionManager.CreateSession(ctx, interviewID, userID); err != nil {
		return nil, fmt.Errorf("[interview] create session: %w", err)
	}

	// 初始化 InterviewState（含 Direction）
	state := &domain.InterviewState{
		InterviewID:  interviewID,
		Stage:        domain.StageIntro,
		Direction:    cfg.Direction,
		Position:     cfg.Position,
		ReportStatus: "pending",
		StartedAt:    time.Now(),
	}
	if err := s.redisCli.SaveInterviewState(ctx, state, s.stateTTL); err != nil {
		return nil, fmt.Errorf("[interview] save state: %w", err)
	}

	return &InterviewCreateResult{
		InterviewID: interviewID,
		Stage:       domain.StageIntro,
		CreatedAt:   state.StartedAt.Format(time.RFC3339),
	}, nil
}

// ProcessAudio 处理音频输入
func (s *interviewServiceImpl) ProcessAudio(ctx context.Context, req AudioRequest) error {
	session, err := s.sessionManager.GetSession(ctx, req.InterviewID)
	if err != nil {
		return fmt.Errorf("[interview] get session: %w", err)
	}

	graphContext, err := s.sessionManager.GetGraphContext(ctx, req.InterviewID)
	if err != nil {
		return fmt.Errorf("[interview] get graph context: %w", err)
	}

	// 注入面试方向到 graph context，供 question_selector 使用
	state, err := s.redisCli.GetInterviewState(ctx, req.InterviewID)
	if err != nil {
		return fmt.Errorf("[interview] get state: %w", err)
	}
	if state != nil {
		graphContext["direction"] = state.Direction
		graphContext["interview_id"] = req.InterviewID
	}

	output, err := s.graph.Invoke(ctx, compose.GraphInput{
		AudioData:   req.AudioData,
		Text:        "",
		InterviewID: req.InterviewID,
		Stage:       session.Stage,
		Context:     graphContext,
	})
	if err != nil {
		return fmt.Errorf("[interview] invoke graph: %w", err)
	}

	userInput := "[音频输入]"
	if err := s.sessionManager.UpdateFromGraphOutput(
		ctx,
		req.InterviewID,
		userInput,
		output.Text,
		output.NewStage,
		output.Context,
	); err != nil {
		return fmt.Errorf("[interview] update session: %w", err)
	}

	// 更新 InterviewState 阶段
	if state != nil && output.NewStage != "" && output.NewStage != state.Stage {
		state.Stage = output.NewStage
		if err := s.redisCli.SaveInterviewState(ctx, state, s.stateTTL); err != nil {
			log.Warnf("[InterviewService] save state after stage change interview_id=%s: %v", req.InterviewID, err)
		}
	}

	return nil
}

// Finish 结束面试
func (s *interviewServiceImpl) Finish(ctx context.Context, interviewID string) (*InterviewFinishResult, error) {
	session, err := s.sessionManager.GetSession(ctx, interviewID)
	if err != nil {
		return nil, fmt.Errorf("[interview] get session: %w", err)
	}

	duration := time.Since(session.CreatedAt)

	if err := s.sessionManager.UpdateStage(ctx, interviewID, domain.StageClosing); err != nil {
		return nil, fmt.Errorf("[interview] update stage: %w", err)
	}

	// 更新 state 的 ReportStatus
	state, err := s.redisCli.GetInterviewState(ctx, interviewID)
	if err == nil && state != nil {
		state.Stage = domain.StageClosing
		state.ReportStatus = "pending"
		if err := s.redisCli.SaveInterviewState(ctx, state, s.stateTTL); err != nil {
			log.Warnf("[InterviewService] save state on finish interview_id=%s: %v", interviewID, err)
		}
	}

	// TODO: 发布 interview_finished 事件到 MQ

	return &InterviewFinishResult{
		InterviewID:     interviewID,
		FinishedAt:      time.Now().Format(time.RFC3339),
		DurationSeconds: int64(duration.Seconds()),
	}, nil
}

// GetState 查询面试当前状态
func (s *interviewServiceImpl) GetState(ctx context.Context, interviewID string) (*domain.InterviewState, error) {
	state, err := s.redisCli.GetInterviewState(ctx, interviewID)
	if err != nil {
		return nil, fmt.Errorf("[interview] get state: %w", err)
	}
	if state != nil {
		return state, nil
	}

	// 降级：从 session 构建
	session, err := s.sessionManager.GetSession(ctx, interviewID)
	if err != nil {
		return nil, fmt.Errorf("[interview] get session: %w", err)
	}

	return &domain.InterviewState{
		InterviewID:    session.InterviewID,
		Stage:          session.Stage,
		QuestionsAsked: session.Stats.QuestionCount,
		StartedAt:      session.CreatedAt,
	}, nil
}

// SubmitCode 提交代码
func (s *interviewServiceImpl) SubmitCode(ctx context.Context, req CodeSubmitRequest) error {
	session, err := s.sessionManager.GetSession(ctx, req.InterviewID)
	if err != nil {
		return fmt.Errorf("[interview] get session: %w", err)
	}

	if session.Stage != domain.StageAlgorithm {
		return fmt.Errorf("[interview] invalid stage for code submission: %s", session.Stage)
	}

	graphContext, err := s.sessionManager.GetGraphContext(ctx, req.InterviewID)
	if err != nil {
		return fmt.Errorf("[interview] get graph context: %w", err)
	}

	codeSubmitText := fmt.Sprintf("我提交了代码：\n```%s\n%s\n```", req.Language, req.Code)
	output, err := s.graph.Invoke(ctx, compose.GraphInput{
		Text:        codeSubmitText,
		InterviewID: req.InterviewID,
		Stage:       session.Stage,
		Context:     graphContext,
	})
	if err != nil {
		return fmt.Errorf("[interview] invoke graph: %w", err)
	}

	if err := s.sessionManager.UpdateFromGraphOutput(
		ctx,
		req.InterviewID,
		codeSubmitText,
		output.Text,
		output.NewStage,
		output.Context,
	); err != nil {
		return fmt.Errorf("[interview] update session: %w", err)
	}

	if err := s.sessionManager.IncrementAlgorithmCount(ctx, req.InterviewID); err != nil {
		return fmt.Errorf("[interview] increment algorithm count: %w", err)
	}

	return nil
}
