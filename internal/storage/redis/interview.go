package redis

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"

	"ai_interview/internal/domain"
)

// SaveInterviewState 保存面试状态到 Redis。
// key: interview:{interview_id}:state
func (c *Client) SaveInterviewState(ctx context.Context, state *domain.InterviewState, ttl time.Duration) error {
	data, err := json.Marshal(state)
	if err != nil {
		return fmt.Errorf("marshal interview state: %w", err)
	}
	return c.rdb.Set(ctx, InterviewStateKey(state.InterviewID), data, ttl).Err()
}

// GetInterviewState 从 Redis 读取面试状态。
// key 不存在时返回 nil, nil。
func (c *Client) GetInterviewState(ctx context.Context, interviewID string) (*domain.InterviewState, error) {
	data, err := c.rdb.Get(ctx, InterviewStateKey(interviewID)).Bytes()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return nil, nil
		}
		return nil, fmt.Errorf("get interview state: %w", err)
	}

	var state domain.InterviewState
	if err := json.Unmarshal(data, &state); err != nil {
		return nil, fmt.Errorf("unmarshal interview state: %w", err)
	}
	return &state, nil
}

// DeleteInterviewState 删除面试状态（面试结束时调用）。
func (c *Client) DeleteInterviewState(ctx context.Context, interviewID string) error {
	return c.rdb.Del(ctx, InterviewStateKey(interviewID)).Err()
}
