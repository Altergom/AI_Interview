package handler

import (
	"context"
	"net/http"

	"github.com/cloudwego/hertz/pkg/app"

	"ai_interview/internal/log"
	biz "ai_interview/internal/utils/respx"
)

// ─── 响应结构 ──────────────────────────────────────────────────────────────────

// Result 统一响应格式。
//
//	成功：{"success":true,  "data":{...},  "error":null}
//	失败：{"success":false, "data":null,   "error":{"code":1400,"message":"..."}}
type Result[T any] struct {
	Success bool     `json:"success"`
	Data    T        `json:"data"`
	Error   *ErrBody `json:"error"`
}

// ErrBody 错误体，仅在 success=false 时非 null。
type ErrBody struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

// ─── 辅助函数 ──────────────────────────────────────────────────────────────────

// OK 写成功响应（HTTP 200）。
func OK[T any](ctx context.Context, c *app.RequestContext, data T) {
	c.JSON(http.StatusOK, Result[T]{
		Success: true,
		Data:    data,
		Error:   nil,
	})
}

// Fail 写失败响应，httpStatus 由调用方决定。
func Fail(ctx context.Context, c *app.RequestContext, httpStatus int, code biz.ErrorCode, msg string) {
	c.JSON(httpStatus, Result[any]{
		Success: false,
		Data:    nil,
		Error:   &ErrBody{Code: int(code), Message: msg},
	})
}

// HandleErr 将 error 自动转换为响应：
//   - *BizError → 对应 HTTP 状态码 + 错误码
//   - 其他 error → 500 Internal
//
// 调用后应直接 return，不再写其他响应。
func HandleErr(ctx context.Context, c *app.RequestContext, err error) {
	if be, ok := biz.IsBizError(err); ok {
		c.JSON(bizHTTPStatus(be.Code), Result[any]{
			Success: false,
			Data:    nil,
			Error:   &ErrBody{Code: int(be.Code), Message: be.Message},
		})
		return
	}
	// 非业务错误，记日志，不暴露内部信息
	log.Errorf("[handler] internal error: %v", err)
	c.JSON(http.StatusInternalServerError, Result[any]{
		Success: false,
		Data:    nil,
		Error:   &ErrBody{Code: int(biz.CodeInternal), Message: biz.CodeInternal.Message()},
	})
}

// bizHTTPStatus 将 ErrorCode 映射到 HTTP 状态码。
func bizHTTPStatus(code biz.ErrorCode) int {
	switch code {
	case biz.CodeUnauthorized, biz.CodeWrongPassword:
		return http.StatusUnauthorized
	case biz.CodeForbidden, biz.CodeInterviewForbidden:
		return http.StatusForbidden
	case biz.CodeNotFound,
		biz.CodeResumeNotFound,
		biz.CodeInterviewSessionNotFound,
		biz.CodeInterviewTurnNotFound,
		biz.CodeStorageNotFound,
		biz.CodeKnowledgeBaseNotFound,
		biz.CodeInterviewScheduleNotFound,
		biz.CodeVoiceSessionNotFound,
		biz.CodeAIProviderNotFound:
		return http.StatusNotFound
	case biz.CodeRateLimitExceeded:
		return http.StatusTooManyRequests
	default:
		// 400 兜底：包含 BadRequest、业务校验失败等
		return http.StatusBadRequest
	}
}
