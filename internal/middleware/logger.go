package middleware

import (
	"context"
	"time"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/google/uuid"

	"ai_interview/internal/log"
)

const (
	// CtxKeyRequestID request_id 在 Hertz context 中的键。
	// 业务层通过 c.GetString(CtxKeyRequestID) 读取。
	CtxKeyRequestID = "request_id"

	// HeaderRequestID 客户端可通过此 Header 传入自定义 request_id（链路追踪）。
	HeaderRequestID = "X-Request-ID"
)

// Logger 全局访问日志 + request_id 注入中间件。
//
// 行为：
//  1. 读取 X-Request-ID 请求头，不存在则生成 UUID v4
//  2. 将 request_id 写入 Hertz context（供后续 handler/service 使用）
//  3. 将 request_id 写回响应头，方便前端/网关关联日志
//  4. 请求处理完后打印访问日志（method / path / status / latency / user_id）
//
// 必须注册在 HAuth 之前，否则 user_id 无法读取（HAuth 尚未写入）。
// 实际顺序：Logger → HAuth → handler。
// user_id 在 logger 读取时可能为空（未鉴权路由），属正常现象。
func Logger() app.HandlerFunc {
	return func(ctx context.Context, c *app.RequestContext) {
		start := time.Now()

		// 1. request_id：优先读客户端传入，否则自动生成
		reqID := string(c.GetHeader(HeaderRequestID))
		if reqID == "" {
			reqID = uuid.New().String()
		}
		c.Set(CtxKeyRequestID, reqID)
		c.Header(HeaderRequestID, reqID)

		// 2. 执行后续处理链
		c.Next(ctx)

		// 3. 请求结束后打印访问日志
		// user_id 由 HAuth 注入，未鉴权路由为空字符串
		userID, _ := c.Get("user_id")
		uid, _ := userID.(string)

		log.Infof("[access] %s %s status=%d latency=%s request_id=%s user_id=%s",
			string(c.Method()),
			string(c.Path()),
			c.Response.StatusCode(),
			time.Since(start).String(),
			reqID,
			uid,
		)
	}
}

// GetRequestID 从 Hertz context 中读取 request_id。
func GetRequestID(c *app.RequestContext) string {
	v, _ := c.Get(CtxKeyRequestID)
	id, _ := v.(string)
	return id
}
