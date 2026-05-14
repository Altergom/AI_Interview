package redis

import (
	"context"
	"crypto/sha256"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"

	"ai_interview/internal/domain"
)

// InterviewConfig 面试配置，由 SetConfig 写入。
type InterviewConfig struct {
	Direction string `json:"direction"` // go-backend / java-backend / frontend / algorithm / ai-agent
	Position  string `json:"position"`  // 岗位名称（可选）
}

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

// SaveInterviewConfig 保存面试配置（方向、岗位）。
// key: interview:{interview_id}:config，TTL 与面试状态对齐。
func (c *Client) SaveInterviewConfig(ctx context.Context, interviewID string, cfg InterviewConfig, ttl time.Duration) error {
	data, err := json.Marshal(cfg)
	if err != nil {
		return fmt.Errorf("marshal interview config: %w", err)
	}
	return c.rdb.Set(ctx, InterviewConfigKey(interviewID), data, ttl).Err()
}

// GetInterviewConfig 读取面试配置。key 不存在时返回 nil, nil。
func (c *Client) GetInterviewConfig(ctx context.Context, interviewID string) (*InterviewConfig, error) {
	data, err := c.rdb.Get(ctx, InterviewConfigKey(interviewID)).Bytes()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return nil, nil
		}
		return nil, fmt.Errorf("get interview config: %w", err)
	}
	var cfg InterviewConfig
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("unmarshal interview config: %w", err)
	}
	return &cfg, nil
}

// SaveSession 保存面试会话到 Redis。
// key: interview:session:{interview_id}
func (c *Client) SaveSession(ctx context.Context, session *domain.InterviewSession, ttl time.Duration) error {
	data, err := json.Marshal(session)
	if err != nil {
		return fmt.Errorf("marshal session: %w", err)
	}
	key := fmt.Sprintf("interview:session:%s", session.InterviewID)
	return c.rdb.Set(ctx, key, data, ttl).Err()
}

// GetSession 从 Redis 读取面试会话，key 不存在时返回 nil, nil。
func (c *Client) GetSession(ctx context.Context, interviewID string) (*domain.InterviewSession, error) {
	key := fmt.Sprintf("interview:session:%s", interviewID)
	data, err := c.rdb.Get(ctx, key).Bytes()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return nil, nil
		}
		return nil, fmt.Errorf("get session: %w", err)
	}
	var session domain.InterviewSession
	if err := json.Unmarshal(data, &session); err != nil {
		return nil, fmt.Errorf("unmarshal session: %w", err)
	}
	return &session, nil
}

// DeleteSession 删除面试会话。
func (c *Client) DeleteSession(ctx context.Context, interviewID string) error {
	key := fmt.Sprintf("interview:session:%s", interviewID)
	return c.rdb.Del(ctx, key).Err()
}

// ExpireSession 刷新面试会话 TTL。
func (c *Client) ExpireSession(ctx context.Context, interviewID string, ttl time.Duration) error {
	key := fmt.Sprintf("interview:session:%s", interviewID)
	return c.rdb.Expire(ctx, key, ttl).Err()
}

// questionHash 计算题目文本的去重 hash（SHA-256 前 16 字节 hex）。
func questionHash(question string) string {
	sum := sha256.Sum256([]byte(question))
	return fmt.Sprintf("%x", sum[:16])
}

// MarkQuestionAsked 将题目 hash 写入 Redis Set，随 interview 状态 TTL 自动过期。
func (c *Client) MarkQuestionAsked(ctx context.Context, interviewID, question string, ttl time.Duration) error {
	key := InterviewAskedQuestionsKey(interviewID)
	hash := questionHash(question)
	pipe := c.rdb.Pipeline()
	pipe.SAdd(ctx, key, hash)
	pipe.Expire(ctx, key, ttl)
	_, err := pipe.Exec(ctx)
	if err != nil {
		return fmt.Errorf("mark asked question: %w", err)
	}
	return nil
}

// IsQuestionAsked 检查题目是否已出过。
func (c *Client) IsQuestionAsked(ctx context.Context, interviewID, question string) (bool, error) {
	key := InterviewAskedQuestionsKey(interviewID)
	hash := questionHash(question)
	exists, err := c.rdb.SIsMember(ctx, key, hash).Result()
	if err != nil {
		return false, fmt.Errorf("check asked question: %w", err)
	}
	return exists, nil
}
