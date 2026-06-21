package session

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"

	"gateway/domain"
)

// Manager 负责 GatewaySession 的路由、创建和查询。
// 不承载面试业务判断，只管会话归属和状态存储。
type Manager struct {
	// TODO: 注入 Redis client 用于持久化 GatewaySession
}

func NewManager() *Manager {
	return &Manager{}
}

// Route 按 channel + peerID 查找活跃会话。
// 未命中返回 nil, false，由上层决定是否进入 onboarding。
func (m *Manager) Route(ctx context.Context, channel, peerID string) (*domain.GatewaySession, bool, error) {
	// TODO: 从 Redis 按 key = "gw:session:{channel}:{peerID}" 查活跃 session
	return nil, false, nil
}

// Create 创建新的 GatewaySession。
func (m *Manager) Create(ctx context.Context, channel, peerID, candidateID string) (*domain.GatewaySession, error) {
	now := time.Now()
	s := &domain.GatewaySession{
		SessionID:    uuid.New().String(),
		CandidateID:  candidateID,
		Channel:      channel,
		PeerID:       peerID,
		Status:       domain.StatusNew,
		CreatedAt:    now,
		UpdatedAt:    now,
		LastActiveAt: now,
	}
	// TODO: 持久化到 Redis
	return s, nil
}

// Get 按 session_id 查询会话。
func (m *Manager) Get(ctx context.Context, sessionID string) (*domain.GatewaySession, error) {
	// TODO: 从 Redis 按 key = "gw:session:id:{sessionID}" 查询
	return nil, fmt.Errorf("not implemented")
}

// List 按 candidate_id 分页查询会话列表。
func (m *Manager) List(ctx context.Context, candidateID string, page, pageSize int) ([]*domain.GatewaySession, int64, error) {
	// TODO: 从存储查询
	return nil, 0, nil
}

// UpdateStatus 更新会话状态。
func (m *Manager) UpdateStatus(ctx context.Context, sessionID string, status domain.SessionStatus) error {
	// TODO: 更新 Redis + 持久化存储
	return nil
}
