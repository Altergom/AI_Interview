package service

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"

	"ai_interview/internal/domain"
	"ai_interview/internal/einocore/agent"
	"ai_interview/internal/einocore/compose"
)

// interviewServiceImpl InterviewService 的实现
type interviewServiceImpl struct {
	sessionManager *SessionManager
	graph          *compose.InterviewGraph
}

// NewInterviewService 创建 InterviewService 实例
func NewInterviewService(
	sessionManager *SessionManager,
) (InterviewService, error) {
	// 创建 Supervisor
	supervisor, err := agent.NewSupervisor()
	if err != nil {
		return nil, fmt.Errorf("failed to create supervisor: %w", err)
	}

	// 创建 Graph
	graph, err := compose.NewInterviewGraph(context.Background(), supervisor)
	if err != nil {
		return nil, fmt.Errorf("failed to create graph: %w", err)
	}

	return &interviewServiceImpl{
		sessionManager: sessionManager,
		graph:          graph,
	}, nil
}

// SetConfig 保存面试岗位和方向配置
func (s *interviewServiceImpl) SetConfig(ctx context.Context, req InterviewConfigRequest) (string, error) {
	// TODO: 实现配置保存到 Redis
	// 生成配置 ID
	configID := uuid.New().String()

	// 保存配置（简化实现）
	// 实际应该保存到 Redis，格式：interview:config:{configID}

	return configID, nil
}

// Create 创建面试会话
func (s *interviewServiceImpl) Create(ctx context.Context, userID string) (*InterviewCreateResult, error) {
	// 生成面试 ID
	interviewID := uuid.New().String()

	// 创建会话
	if err := s.sessionManager.CreateSession(ctx, interviewID, userID); err != nil {
		return nil, fmt.Errorf("failed to create session: %w", err)
	}

	return &InterviewCreateResult{
		InterviewID: interviewID,
		Stage:       domain.StageIntro,
		CreatedAt:   time.Now().Format(time.RFC3339),
	}, nil
}

// ProcessAudio 处理音频输入
func (s *interviewServiceImpl) ProcessAudio(ctx context.Context, req AudioRequest) error {
	// 1. 获取当前会话状态
	session, err := s.sessionManager.GetSession(ctx, req.InterviewID)
	if err != nil {
		return fmt.Errorf("failed to get session: %w", err)
	}

	// 2. 获取 Graph 上下文
	graphContext, err := s.sessionManager.GetGraphContext(ctx, req.InterviewID)
	if err != nil {
		return fmt.Errorf("failed to get graph context: %w", err)
	}

	// 3. 调用 Graph 处理
	output, err := s.graph.Invoke(ctx, compose.GraphInput{
		AudioData:   req.AudioData,
		Text:        "", // 音频输入时文本为空
		InterviewID: req.InterviewID,
		Stage:       session.Stage,
		Context:     graphContext,
	})
	if err != nil {
		return fmt.Errorf("failed to invoke graph: %w", err)
	}

	// 4. 更新会话
	// 注意：音频输入时，用户消息内容由 ASR 转换后的文本
	userInput := "[音频输入]" // 简化处理，实际应该从 Graph 输出中获取 ASR 结果
	if err := s.sessionManager.UpdateFromGraphOutput(
		ctx,
		req.InterviewID,
		userInput,
		output.Text,
		output.NewStage,
		output.Context,
	); err != nil {
		return fmt.Errorf("failed to update session: %w", err)
	}

	// 5. TODO: 推送 AI 回复到前端（通过 SSE）
	// 这里需要一个事件推送机制

	return nil
}

// Finish 结束面试
func (s *interviewServiceImpl) Finish(ctx context.Context, interviewID string) (*InterviewFinishResult, error) {
	// 1. 获取会话
	session, err := s.sessionManager.GetSession(ctx, interviewID)
	if err != nil {
		return nil, fmt.Errorf("failed to get session: %w", err)
	}

	// 2. 计算面试时长
	duration := time.Since(session.CreatedAt)

	// 3. 更新阶段为结束
	if err := s.sessionManager.UpdateStage(ctx, interviewID, domain.StageClosing); err != nil {
		return nil, fmt.Errorf("failed to update stage: %w", err)
	}

	// 4. TODO: 发布 interview_finished 事件到 MQ

	return &InterviewFinishResult{
		InterviewID:     interviewID,
		FinishedAt:      time.Now().Format(time.RFC3339),
		DurationSeconds: int64(duration.Seconds()),
	}, nil
}

// GetState 查询面试当前状态
func (s *interviewServiceImpl) GetState(ctx context.Context, interviewID string) (*domain.InterviewState, error) {
	// 获取会话
	session, err := s.sessionManager.GetSession(ctx, interviewID)
	if err != nil {
		return nil, fmt.Errorf("failed to get session: %w", err)
	}

	// 构建状态
	state := &domain.InterviewState{
		InterviewID:              session.InterviewID,
		Stage:                    session.Stage,
		QuestionsAsked:           session.Stats.QuestionCount,
		CurrentQuestionFollowups: 0, // TODO: 从上下文中获取
		StartedAt:                session.CreatedAt,
	}

	return state, nil
}

// SubmitCode 提交代码
func (s *interviewServiceImpl) SubmitCode(ctx context.Context, req CodeSubmitRequest) error {
	// 1. 获取当前会话状态
	session, err := s.sessionManager.GetSession(ctx, req.InterviewID)
	if err != nil {
		return fmt.Errorf("failed to get session: %w", err)
	}

	// 2. 验证当前阶段（应该在算法题阶段）
	if session.Stage != domain.StageAlgorithm {
		return fmt.Errorf("invalid stage for code submission: %s", session.Stage)
	}

	// 3. TODO: 调用 Code Judge Tool 判断代码正确性

	// 4. 获取 Graph 上下文
	graphContext, err := s.sessionManager.GetGraphContext(ctx, req.InterviewID)
	if err != nil {
		return fmt.Errorf("failed to get graph context: %w", err)
	}

	// 5. 调用 Graph 处理（传递代码提交信息）
	codeSubmitText := fmt.Sprintf("我提交了代码：\n```%s\n%s\n```", req.Language, req.Code)
	output, err := s.graph.Invoke(ctx, compose.GraphInput{
		AudioData:   nil,
		Text:        codeSubmitText,
		InterviewID: req.InterviewID,
		Stage:       session.Stage,
		Context:     graphContext,
	})
	if err != nil {
		return fmt.Errorf("failed to invoke graph: %w", err)
	}

	// 6. 更新会话
	if err := s.sessionManager.UpdateFromGraphOutput(
		ctx,
		req.InterviewID,
		codeSubmitText,
		output.Text,
		output.NewStage,
		output.Context,
	); err != nil {
		return fmt.Errorf("failed to update session: %w", err)
	}

	// 7. 增加算法题计数
	if err := s.sessionManager.IncrementAlgorithmCount(ctx, req.InterviewID); err != nil {
		return fmt.Errorf("failed to increment algorithm count: %w", err)
	}

	return nil
}

// ProcessText 处理文本输入（辅助方法）
// 用于测试或文本聊天场景
func (s *interviewServiceImpl) ProcessText(ctx context.Context, interviewID, text string) (string, error) {
	// 1. 获取当前会话状态
	session, err := s.sessionManager.GetSession(ctx, interviewID)
	if err != nil {
		return "", fmt.Errorf("failed to get session: %w", err)
	}

	// 2. 获取 Graph 上下文
	graphContext, err := s.sessionManager.GetGraphContext(ctx, interviewID)
	if err != nil {
		return "", fmt.Errorf("failed to get graph context: %w", err)
	}

	// 3. 调用 Graph 处理
	output, err := s.graph.Invoke(ctx, compose.GraphInput{
		AudioData:   nil,
		Text:        text,
		InterviewID: interviewID,
		Stage:       session.Stage,
		Context:     graphContext,
	})
	if err != nil {
		return "", fmt.Errorf("failed to invoke graph: %w", err)
	}

	// 4. 更新会话
	if err := s.sessionManager.UpdateFromGraphOutput(
		ctx,
		interviewID,
		text,
		output.Text,
		output.NewStage,
		output.Context,
	); err != nil {
		return "", fmt.Errorf("failed to update session: %w", err)
	}

	return output.Text, nil
}
