package ratelimit

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/redis/go-redis/v9"

	biz "ai_interview/internal/errors"
	"ai_interview/internal/log"
)

// Dimension 限流维度。
type Dimension string

const (
	DimensionIP   Dimension = "ip"
	DimensionUser Dimension = "user"
)

// Config 单个限流规则。
type Config struct {
	// Limit 窗口内最大请求数
	Limit int
	// Window 滑动窗口大小
	Window time.Duration
}

// defaultConfigs 各维度默认限流配置，与 TODO.md 约定一致。
var defaultConfigs = map[Dimension]Config{
	DimensionIP:   {Limit: 10, Window: time.Minute},
	DimensionUser: {Limit: 30, Window: time.Minute},
}

// slidingWindowLua 滑动窗口限流 Lua 脚本（原子操作，防竞争）。
//
// KEYS[1] = ratelimit key
// ARGV[1] = 当前时间戳（毫秒）
// ARGV[2] = 窗口大小（毫秒）
// ARGV[3] = 窗口内最大请求数
// ARGV[4] = key TTL（秒，= 窗口大小 + 1s 缓冲）
//
// 返回：
//
//	0 — 允许通过
//	1 — 触发限流
var slidingWindowLua = redis.NewScript(`
local key    = KEYS[1]
local now    = tonumber(ARGV[1])
local window = tonumber(ARGV[2])
local limit  = tonumber(ARGV[3])
local ttl    = tonumber(ARGV[4])

-- 移除窗口外的旧记录
redis.call('ZREMRANGEBYSCORE', key, 0, now - window)

-- 统计窗口内请求数
local count = redis.call('ZCARD', key)

if count >= limit then
    return 1
end

-- 记录本次请求（score = 时间戳，member = 时间戳，保证唯一性用随机后缀）
redis.call('ZADD', key, now, now .. '-' .. math.random(1000000))
redis.call('EXPIRE', key, ttl)
return 0
`)

// Limiter 持有 Redis 客户端，执行限流判断。
// 导出供 WebSocket handler 等非中间件场景直接调用。
type Limiter struct {
	rdb     *redis.Client
	configs map[Dimension]Config
}

// NewLimiter 创建 Limiter，overrides 为空则使用默认配置。
func NewLimiter(rdb *redis.Client, overrides map[Dimension]Config) *Limiter {
	configs := make(map[Dimension]Config)
	for k, v := range defaultConfigs {
		configs[k] = v
	}
	for k, v := range overrides {
		configs[k] = v
	}
	return &Limiter{rdb: rdb, configs: configs}
}

// rateLimiter 内部使用，复用 Limiter 逻辑。
type rateLimiter = Limiter

func newRateLimiter(rdb *redis.Client, overrides map[Dimension]Config) *rateLimiter {
	return NewLimiter(rdb, overrides)
}

// Allow 返回 true 表示允许通过，false 表示触发限流（导出版本）。
func (r *Limiter) Allow(ctx context.Context, handler string, dim Dimension, value string) bool {
	return r.allow(ctx, handler, dim, value)
}

// allow 返回 true 表示允许通过，false 表示触发限流。
func (r *Limiter) allow(ctx context.Context, handler string, dim Dimension, value string) bool {
	cfg, ok := r.configs[dim]
	if !ok {
		return true // 未配置的维度放行
	}

	key := fmt.Sprintf("ratelimit:%s:%s:%s", handler, dim, value)
	nowMs := time.Now().UnixMilli()
	windowMs := cfg.Window.Milliseconds()
	ttlSec := int(cfg.Window.Seconds()) + 1

	result, err := slidingWindowLua.Run(ctx, r.rdb, []string{key},
		nowMs, windowMs, cfg.Limit, ttlSec,
	).Int()
	if err != nil {
		// Redis 故障时放行，避免限流组件导致全局不可用
		log.Warnf("[ratelimit] lua error, fail-open: %v", err)
		return true
	}
	return result == 0
}

// Middleware 返回 Hertz 限流中间件。
//
// handler 用于区分不同端点的限流桶，例如 "resume.parse"。
// dims 指定启用的维度，为空则同时启用 IP 和 USER。
//
// 用法：
//
//	resume.POST("/parse", ratelimit.Middleware(rdb, "resume.parse"), h.resume.Parse)
func Middleware(rdb *redis.Client, handler string, dims ...Dimension) app.HandlerFunc {
	if len(dims) == 0 {
		dims = []Dimension{DimensionIP, DimensionUser}
	}
	rl := newRateLimiter(rdb, nil)

	return func(ctx context.Context, c *app.RequestContext) {
		for _, dim := range dims {
			var value string
			switch dim {
			case DimensionIP:
				value = clientIP(c)
			case DimensionUser:
				value = userID(c)
				if value == "" {
					continue // 未登录用户跳过 USER 维度
				}
			}

			if !rl.allow(ctx, handler, dim, value) {
				log.Infof("[ratelimit] blocked handler=%s dim=%s value=%s", handler, dim, value)
				c.JSON(http.StatusTooManyRequests, map[string]any{
					"success": false,
					"data":    nil,
					"error": map[string]any{
						"code":    int(biz.CodeRateLimitExceeded),
						"message": biz.CodeRateLimitExceeded.Message(),
					},
				})
				c.Abort()
				return
			}
		}
		c.Next(ctx)
	}
}

// clientIP 获取真实客户端 IP，优先读 X-Forwarded-For。
func clientIP(c *app.RequestContext) string {
	if xff := string(c.GetHeader("X-Forwarded-For")); xff != "" {
		// 取第一个 IP（最左侧为原始客户端）
		if idx := strings.Index(xff, ","); idx > 0 {
			return strings.TrimSpace(xff[:idx])
		}
		return strings.TrimSpace(xff)
	}
	if xri := string(c.GetHeader("X-Real-IP")); xri != "" {
		return strings.TrimSpace(xri)
	}
	return c.ClientIP()
}

// userID 从 context 中读取已由 JWT 中间件注入的 user_id。
func userID(c *app.RequestContext) string {
	v, _ := c.Get("user_id")
	id, _ := v.(string)
	return id
}
