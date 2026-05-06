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

// SaveResume 保存结构化简历到 Redis。
// key: resume:{user_id}
func (c *Client) SaveResume(ctx context.Context, resume *domain.StructuredResume, ttl time.Duration) error {
	data, err := json.Marshal(resume)
	if err != nil {
		return fmt.Errorf("marshal resume: %w", err)
	}
	return c.rdb.Set(ctx, ResumeKey(resume.UserID), data, ttl).Err()
}

// GetResume 从 Redis 读取结构化简历。
// key 不存在时返回 nil, nil。
func (c *Client) GetResume(ctx context.Context, userID string) (*domain.StructuredResume, error) {
	data, err := c.rdb.Get(ctx, ResumeKey(userID)).Bytes()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return nil, nil
		}
		return nil, fmt.Errorf("get resume: %w", err)
	}

	var resume domain.StructuredResume
	if err := json.Unmarshal(data, &resume); err != nil {
		return nil, fmt.Errorf("unmarshal resume: %w", err)
	}
	return &resume, nil
}

// DeleteResume 删除简历缓存。
func (c *Client) DeleteResume(ctx context.Context, userID string) error {
	return c.rdb.Del(ctx, ResumeKey(userID)).Err()
}
