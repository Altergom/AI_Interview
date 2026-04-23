package redis

import (
	"context"
	"time"

	"ai_interview/internal/domain"
)

// SaveResume 保存结构化简历到 Redis。
// key: resume:{user_id}
func (c *Client) SaveResume(ctx context.Context, resume *domain.StructuredResume, ttl time.Duration) error {
	// TODO: 实现简历保存
	// 1. 序列化 resume 为 JSON
	// 2. 使用 keys.ResumeKey() 生成 key
	// 3. SET key value EX ttl
	return nil
}

// GetResume 从 Redis 读取结构化简历。
func (c *Client) GetResume(ctx context.Context, userID string) (*domain.StructuredResume, error) {
	// TODO: 实现简历读取
	// 1. 使用 keys.ResumeKey() 生成 key
	// 2. GET key
	// 3. 反序列化 JSON 为 domain.StructuredResume
	return nil, nil
}

// DeleteResume 删除简历缓存。
func (c *Client) DeleteResume(ctx context.Context, userID string) error {
	// TODO: 实现删除
	return nil
}
