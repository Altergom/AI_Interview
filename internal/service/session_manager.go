package service

import (
	"context"
	"fmt"
	"time"

	"ai_interview/internal/domain"
	redistore "ai_interview/internal/storage/redis"
)

// SessionManager 面试会话管理器
type SessionManager struct {
	rdb *redistore.Client
	ttl time.Duration
}

// NewSessionManager 创建会话管理器
func NewSessionManager(rdb *redistore.Client, ttl time.Duration) *SessionManager {
	return &SessionManager{rdb: rdb, ttl: ttl}
}

// CreateSession 创建新的面试会话
func (sm *SessionManager) CreateSession(ctx context.Context, interviewID, userID string) error {
	session := &domain.InterviewSession{
		InterviewID: interviewID,
		UserID:      userID,
		Stage:       domain.StageIntro,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
		History:     make([]domain.SessionMessage, 0),
		Stats:       domain.SessionStats{},
		Context:     make(map[string]any),
	}
	return sm.rdb.SaveSession(ctx, session, sm.ttl)
}

// GetSession 获取面试会话，不存在时返回 error。
func (sm *SessionManager) GetSession(ctx context.Context, interviewID string) (*domain.InterviewSession, error) {
	session, err := sm.rdb.GetSession(ctx, interviewID)
	if err != nil {
		return nil, fmt.Errorf("get session: %w", err)
	}
	if session == nil {
		return nil, fmt.Errorf("session not found: %s", interviewID)
	}
	return session, nil
}

// saveSession 保存会话到 Redis
func (sm *SessionManager) saveSession(ctx context.Context, session *domain.InterviewSession) error {
	return sm.rdb.SaveSession(ctx, session, sm.ttl)
}

// AddMessage 添加对话消息
func (sm *SessionManager) AddMessage(ctx context.Context, interviewID, role, content string) error {
	session, err := sm.GetSession(ctx, interviewID)
	if err != nil {
		return err
	}
	session.History = append(session.History, domain.SessionMessage{
		Role:      role,
		Content:   content,
		Timestamp: time.Now(),
	})
	session.Stats.TotalRounds++
	return sm.saveSession(ctx, session)
}

// GetHistory 获取历史对话
func (sm *SessionManager) GetHistory(ctx context.Context, interviewID string) ([]domain.SessionMessage, error) {
	session, err := sm.GetSession(ctx, interviewID)
	if err != nil {
		return nil, err
	}
	return session.History, nil
}

// UpdateStage 更新面试阶段
func (sm *SessionManager) UpdateStage(ctx context.Context, interviewID string, newStage domain.InterviewStage) error {
	session, err := sm.GetSession(ctx, interviewID)
	if err != nil {
		return err
	}
	session.Stage = newStage
	return sm.saveSession(ctx, session)
}

// GetStage 获取当前阶段
func (sm *SessionManager) GetStage(ctx context.Context, interviewID string) (domain.InterviewStage, error) {
	session, err := sm.GetSession(ctx, interviewID)
	if err != nil {
		return "", err
	}
	return session.Stage, nil
}

// IncrementQuestionCount 增加问题计数
func (sm *SessionManager) IncrementQuestionCount(ctx context.Context, interviewID string) error {
	session, err := sm.GetSession(ctx, interviewID)
	if err != nil {
		return err
	}
	session.Stats.QuestionCount++
	return sm.saveSession(ctx, session)
}

// IncrementAlgorithmCount 增加算法题计数
func (sm *SessionManager) IncrementAlgorithmCount(ctx context.Context, interviewID string) error {
	session, err := sm.GetSession(ctx, interviewID)
	if err != nil {
		return err
	}
	session.Stats.AlgorithmCount++
	return sm.saveSession(ctx, session)
}

// GetStats 获取统计信息
func (sm *SessionManager) GetStats(ctx context.Context, interviewID string) (*domain.SessionStats, error) {
	session, err := sm.GetSession(ctx, interviewID)
	if err != nil {
		return nil, err
	}
	return &session.Stats, nil
}

// UpdateContext 更新上下文信息
func (sm *SessionManager) UpdateContext(ctx context.Context, interviewID string, key string, value any) error {
	session, err := sm.GetSession(ctx, interviewID)
	if err != nil {
		return err
	}
	session.Context[key] = value
	return sm.saveSession(ctx, session)
}

// GetContext 获取上下文信息
func (sm *SessionManager) GetContext(ctx context.Context, interviewID string) (map[string]any, error) {
	session, err := sm.GetSession(ctx, interviewID)
	if err != nil {
		return nil, err
	}
	return session.Context, nil
}

// DeleteSession 删除会话
func (sm *SessionManager) DeleteSession(ctx context.Context, interviewID string) error {
	return sm.rdb.DeleteSession(ctx, interviewID)
}

// ExtendTTL 延长会话过期时间
func (sm *SessionManager) ExtendTTL(ctx context.Context, interviewID string) error {
	return sm.rdb.ExpireSession(ctx, interviewID, sm.ttl)
}

// GetGraphContext 获取 Graph 所需的上下文
func (sm *SessionManager) GetGraphContext(ctx context.Context, interviewID string) (map[string]any, error) {
	session, err := sm.GetSession(ctx, interviewID)
	if err != nil {
		return nil, err
	}

	history := make([]map[string]string, 0, len(session.History))
	for _, msg := range session.History {
		history = append(history, map[string]string{
			"role":    msg.Role,
			"content": msg.Content,
		})
	}

	graphContext := map[string]any{
		"history":         history,
		"question_count":  session.Stats.QuestionCount,
		"algorithm_count": session.Stats.AlgorithmCount,
		"total_rounds":    session.Stats.TotalRounds,
	}
	for k, v := range session.Context {
		graphContext[k] = v
	}
	return graphContext, nil
}

// UpdateFromGraphOutput 根据 Graph 的输出更新会话
func (sm *SessionManager) UpdateFromGraphOutput(
	ctx context.Context,
	interviewID string,
	userInput string,
	aiResponse string,
	newStage domain.InterviewStage,
	graphContext map[string]any,
) error {
	session, err := sm.GetSession(ctx, interviewID)
	if err != nil {
		return err
	}

	now := time.Now()
	session.History = append(session.History,
		domain.SessionMessage{Role: "user", Content: userInput, Timestamp: now},
		domain.SessionMessage{Role: "assistant", Content: aiResponse, Timestamp: now},
	)
	session.Stage = newStage
	session.Stats.TotalRounds++

	for k, v := range graphContext {
		session.Context[k] = v
	}
	return sm.saveSession(ctx, session)
}

// ShouldAdvanceStage 判断是否应该切换阶段
func (sm *SessionManager) ShouldAdvanceStage(ctx context.Context, interviewID string) (bool, error) {
	session, err := sm.GetSession(ctx, interviewID)
	if err != nil {
		return false, err
	}

	switch session.Stage {
	case domain.StageIntro:
		return session.Stats.TotalRounds >= 3, nil
	case domain.StageQuestioning:
		return session.Stats.QuestionCount >= 5, nil
	case domain.StageAlgorithm:
		return session.Stats.AlgorithmCount >= 1, nil
	default:
		return false, nil
	}
}
