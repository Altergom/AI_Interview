package redis

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"

	"ai_interview/internal/log"
)

// Options Redis 连接配置，全部字段均有合理默认值。
type Options struct {
	Addr     string
	Password string
	DB       int

	// 连接池
	PoolSize     int           // 最大连接数，默认 10
	MinIdleConns int           // 最小空闲连接，默认 2
	DialTimeout  time.Duration // 建连超时，默认 5s
	ReadTimeout  time.Duration // 读超时，默认 3s
	WriteTimeout time.Duration // 写超时，默认 3s
}

func (o *Options) withDefaults() {
	if o.PoolSize <= 0 {
		o.PoolSize = 10
	}
	if o.MinIdleConns <= 0 {
		o.MinIdleConns = 2
	}
	if o.DialTimeout <= 0 {
		o.DialTimeout = 5 * time.Second
	}
	if o.ReadTimeout <= 0 {
		o.ReadTimeout = 3 * time.Second
	}
	if o.WriteTimeout <= 0 {
		o.WriteTimeout = 3 * time.Second
	}
}

// Client 封装 Redis 连接。
type Client struct {
	rdb *redis.Client
}

// New 初始化 Redis 客户端并验证连通性。
func New(ctx context.Context, opts Options) (*Client, error) {
	opts.withDefaults()

	rdb := redis.NewClient(&redis.Options{
		Addr:         opts.Addr,
		Password:     opts.Password,
		DB:           opts.DB,
		PoolSize:     opts.PoolSize,
		MinIdleConns: opts.MinIdleConns,
		DialTimeout:  opts.DialTimeout,
		ReadTimeout:  opts.ReadTimeout,
		WriteTimeout: opts.WriteTimeout,
	})

	if err := rdb.Ping(ctx).Err(); err != nil {
		_ = rdb.Close()
		return nil, fmt.Errorf("ping redis: %w", err)
	}

	log.Infof("[Redis] connected addr=%s db=%d poolSize=%d minIdle=%d",
		opts.Addr, opts.DB, opts.PoolSize, opts.MinIdleConns)

	return &Client{rdb: rdb}, nil
}

// Close 关闭 Redis 连接。
func (c *Client) Close() error {
	log.Infof("[Redis] closing connection")
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
