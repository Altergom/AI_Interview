package service

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"

	"ai_interview/internal/domain"
)

// SessionManager 面试会话管理器
// 负责管理面试会话的状态、历史对话、阶段切换等
type SessionManager struct {
	redis *redis.Client
	ttl   time.Duration // 会话过期时间
}

// NewSessionManager 创建会话管理器
func NewSessionManager(redisClient *redis.Client, ttl time.Duration) *SessionManager {
	return &SessionManager{
		redis: redisClient,
		ttl:   ttl,
	}
}

// InterviewSession 面试会话数据
type InterviewSession struct {
	InterviewID string                `json:"interview_id"`
	UserID      string                `json:"user_id"`
	Stage       domain.InterviewStage `json:"stage"`
	CreatedAt   time.Time             `json:"created_at"`
	UpdatedAt   time.Time             `json:"updated_at"`

	// 历史对话
	History []Message `json:"history"`

	// 统计信息
	Stats SessionStats `json:"stats"`

	// 上下文信息
	Context map[string]any `json:"context"`
}

// Message 对话消息
type Message struct {
	Role      string    `json:"role"`      // user | assistant
	Content   string    `json:"content"`   // 消息内容
	Timestamp time.Time `json:"timestamp"` // 时间戳
}

// SessionStats 会话统计信息
type SessionStats struct {
	QuestionCount  int `json:"question_count"`  // 已问问题数
	AlgorithmCount int `json:"algorithm_count"` // 已做算法题数
	TotalRounds    int `json:"total_rounds"`    // 总对话轮数
}

// Redis Key 生成
func (sm *SessionManager) sessionKey(interviewID string) string {
	return fmt.Sprintf("interview:session:%s", interviewID)
}

// CreateSession 创建新的面试会话
func (sm *SessionManager) CreateSession(ctx context.Context, interviewID, userID string) error {
	session := &InterviewSession{
		InterviewID: interviewID,
		UserID:      userID,
		Stage:       domain.StageIntro, // 初始阶段：自我介绍
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
		History:     make([]Message, 0),
		Stats: SessionStats{
			QuestionCount:  0,
			AlgorithmCount: 0,
			TotalRounds:    0,
		},
		Context: make(map[string]any),
	}

	return sm.saveSession(ctx, session)
}

// GetSession 获取面试会话
func (sm *SessionManager) GetSession(ctx context.Context, interviewID string) (*InterviewSession, error) {
	key := sm.sessionKey(interviewID)

	data, err := sm.redis.Get(ctx, key).Bytes()
	if err != nil {
		if err == redis.Nil {
			return nil, fmt.Errorf("session not found: %s", interviewID)
		}
		return nil, fmt.Errorf("failed to get session: %w", err)
	}

	var session InterviewSession
	if err := json.Unmarshal(data, &session); err != nil {
		return nil, fmt.Errorf("failed to unmarshal session: %w", err)
	}

	return &session, nil
}

// saveSession 保存会话到 Redis
func (sm *SessionManager) saveSession(ctx context.Context, session *InterviewSession) error {
	session.UpdatedAt = time.Now()

	data, err := json.Marshal(session)
	if err != nil {
		return fmt.Errorf("failed to marshal session: %w", err)
	}

	key := sm.sessionKey(session.InterviewID)
	if err := sm.redis.Set(ctx, key, data, sm.ttl).Err(); err != nil {
		return fmt.Errorf("failed to save session: %w", err)
	}

	return nil
}

// AddMessage 添加对话消息
func (sm *SessionManager) AddMessage(ctx context.Context, interviewID, role, content string) error {
	session, err := sm.GetSession(ctx, interviewID)
	if err != nil {
		return err
	}

	// 添加消息
	session.History = append(session.History, Message{
		Role:      role,
		Content:   content,
		Timestamp: time.Now(),
	})

	// 更新统计
	session.Stats.TotalRounds++

	return sm.saveSession(ctx, session)
}

// GetHistory 获取历史对话
func (sm *SessionManager) GetHistory(ctx context.Context, interviewID string) ([]Message, error) {
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
func (sm *SessionManager) GetStats(ctx context.Context, interviewID string) (*SessionStats, error) {
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
	key := sm.sessionKey(interviewID)
	return sm.redis.Del(ctx, key).Err()
}

// ExtendTTL 延长会话过期时间
func (sm *SessionManager) ExtendTTL(ctx context.Context, interviewID string) error {
	key := sm.sessionKey(interviewID)
	return sm.redis.Expire(ctx, key, sm.ttl).Err()
}

// GetGraphContext 获取 Graph 所需的上下文
// 返回格式化的历史对话，用于传递给 Graph
func (sm *SessionManager) GetGraphContext(ctx context.Context, interviewID string) (map[string]any, error) {
	session, err := sm.GetSession(ctx, interviewID)
	if err != nil {
		return nil, err
	}

	// 转换历史对话格式
	history := make([]map[string]string, 0, len(session.History))
	for _, msg := range session.History {
		history = append(history, map[string]string{
			"role":    msg.Role,
			"content": msg.Content,
		})
	}

	// 构建上下文
	graphContext := map[string]any{
		"history":         history,
		"question_count":  session.Stats.QuestionCount,
		"algorithm_count": session.Stats.AlgorithmCount,
		"total_rounds":    session.Stats.TotalRounds,
	}

	// 合并自定义上下文
	for k, v := range session.Context {
		graphContext[k] = v
	}

	return graphContext, nil
}

// UpdateFromGraphOutput 根据 Graph 的输出更新会话
// 添加对话历史、更新阶段、更新上下文
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

	// 添加用户消息
	session.History = append(session.History, Message{
		Role:      "user",
		Content:   userInput,
		Timestamp: time.Now(),
	})

	// 添加 AI 回复
	session.History = append(session.History, Message{
		Role:      "assistant",
		Content:   aiResponse,
		Timestamp: time.Now(),
	})

	// 更新阶段
	session.Stage = newStage

	// 更新统计
	session.Stats.TotalRounds++

	// 更新上下文（合并 Graph 返回的上下文）
	for k, v := range graphContext {
		session.Context[k] = v
	}

	return sm.saveSession(ctx, session)
}

// ShouldAdvanceStage 判断是否应该切换阶段（基于规则）
// 这是一个辅助方法，可以作为 Supervisor 判断的补充
func (sm *SessionManager) ShouldAdvanceStage(ctx context.Context, interviewID string) (bool, error) {
	session, err := sm.GetSession(ctx, interviewID)
	if err != nil {
		return false, err
	}

	switch session.Stage {
	case domain.StageIntro:
		// 自我介绍阶段：3-5 轮对话后可以切换
		return session.Stats.TotalRounds >= 3, nil

	case domain.StageQuestioning:
		// 技术问答阶段：5-8 个问题后可以切换
		return session.Stats.QuestionCount >= 5, nil

	case domain.StageAlgorithm:
		// 算法题阶段：1-2 道题后可以切换
		return session.Stats.AlgorithmCount >= 1, nil

	case domain.StageClosing:
		// 反问阶段：不自动切换
		return false, nil

	default:
		return false, nil
	}
}
