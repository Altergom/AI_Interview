package redis

import (
	"context"
	"time"

	"ai_interview/internal/domain"
)

// SaveInterviewState 保存面试状态到 Redis。
// key: interview:{interview_id}:state
func (c *Client) SaveInterviewState(ctx context.Context, state *domain.InterviewState, ttl time.Duration) error {
	// TODO: 实现面试状态保存
	// 1. 序列化 state 为 JSON
	// 2. 使用 keys.InterviewStateKey() 生成 key
	// 3. SET key value EX ttl
	return nil
}

// GetInterviewState 从 Redis 读取面试状态。
func (c *Client) GetInterviewState(ctx context.Context, interviewID string) (*domain.InterviewState, error) {
	// TODO: 实现面试状态读取
	// 1. 使用 keys.InterviewStateKey() 生成 key
	// 2. GET key
	// 3. 反序列化 JSON 为 domain.InterviewState
	return nil, nil
}

// DeleteInterviewState 删除面试状态（面试结束时调用）。
func (c *Client) DeleteInterviewState(ctx context.Context, interviewID string) error {
	// TODO: 实现删除
	return nil
}
