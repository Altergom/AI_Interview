package redis

import (
	"context"
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

// New 初始化 Redis 客户端。
func New(ctx context.Context, opts Options) (*Client, error) {
	// TODO: 实现 Redis 连接初始化
	return nil, nil
}

// Close 关闭 Redis 连接。
func (c *Client) Close() error {
	// TODO: 实现关闭逻辑
	return nil
}

// Ping 检查 Redis 连通性。
func (c *Client) Ping(ctx context.Context) error {
	// TODO: 实现 ping
	return nil
}

// Client 返回底层 *redis.Client，供业务层直接操作。
func (c *Client) Client() *redis.Client {
	return c.rdb
}

// SetWithTTL 通用的 set 方法，带 TTL。
func (c *Client) SetWithTTL(ctx context.Context, key string, value interface{}, ttl time.Duration) error {
	// TODO: 实现 set
	return nil
}

// Get 通用的 get 方法。
func (c *Client) Get(ctx context.Context, key string) (string, error) {
	// TODO: 实现 get
	return "", nil
}

// Del 删除 key。
func (c *Client) Del(ctx context.Context, keys ...string) error {
	// TODO: 实现 del
	return nil
}
