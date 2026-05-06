package redis

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

// Client 封装 Redis 连接。
type Client struct {
	rdb *redis.Client
}

// Options Redis 连接配置。
type Options struct {
	Addr     string
	Password string
	DB       int
}

// New 初始化 Redis 客户端并验证连通性。
func New(ctx context.Context, opts Options) (*Client, error) {
	rdb := redis.NewClient(&redis.Options{
		Addr:     opts.Addr,
		Password: opts.Password,
		DB:       opts.DB,
	})

	if err := rdb.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("ping redis: %w", err)
	}

	return &Client{rdb: rdb}, nil
}

// Close 关闭 Redis 连接。
func (c *Client) Close() error {
	return c.rdb.Close()
}

// Ping 检查 Redis 连通性。
func (c *Client) Ping(ctx context.Context) error {
	return c.rdb.Ping(ctx).Err()
}

// Client 返回底层 *redis.Client，供业务层直接操作。
func (c *Client) Client() *redis.Client {
	return c.rdb
}

// SetWithTTL 通用 set，带 TTL。
func (c *Client) SetWithTTL(ctx context.Context, key string, value any, ttl time.Duration) error {
	return c.rdb.Set(ctx, key, value, ttl).Err()
}

// Get 通用 get，key 不存在时返回 redis.Nil 错误。
func (c *Client) Get(ctx context.Context, key string) (string, error) {
	return c.rdb.Get(ctx, key).Result()
}

// Del 删除一个或多个 key。
func (c *Client) Del(ctx context.Context, keys ...string) error {
	return c.rdb.Del(ctx, keys...).Err()
}
