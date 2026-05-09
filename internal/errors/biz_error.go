package errors

import (
	"errors"
	"fmt"
)

// BizError 业务错误，携带错误码、用户可见消息和内部原因。
// 通过 New / Wrap 构造，禁止直接用 errors.New / fmt.Errorf 在 service / handler 层裸抛。
type BizError struct {
	Code    ErrorCode
	Message string // 用户可见的错误描述（可覆盖 ErrorCode 默认文案）
	Cause   error  // 内部原因，用于日志，不暴露给客户端
}

// New 用默认文案构造 BizError。
func New(code ErrorCode) *BizError {
	return &BizError{Code: code, Message: code.Message()}
}

// NewMsg 用自定义文案构造 BizError。
func NewMsg(code ErrorCode, msg string) *BizError {
	return &BizError{Code: code, Message: msg}
}

// Wrap 包装内部错误（Cause 仅写日志，不返回客户端）。
func Wrap(code ErrorCode, cause error) *BizError {
	return &BizError{Code: code, Message: code.Message(), Cause: cause}
}

// WrapMsg 包装内部错误并自定义用户可见文案。
func WrapMsg(code ErrorCode, msg string, cause error) *BizError {
	return &BizError{Code: code, Message: msg, Cause: cause}
}

// Error 实现 error 接口，输出内部完整信息（含 cause，用于日志）。
func (e *BizError) Error() string {
	if e.Cause != nil {
		return fmt.Sprintf("biz_error[%d]: %s: %v", e.Code, e.Message, e.Cause)
	}
	return fmt.Sprintf("biz_error[%d]: %s", e.Code, e.Message)
}

// Unwrap 支持 errors.Is / errors.As 链式匹配。
func (e *BizError) Unwrap() error {
	return e.Cause
}

// Is 允许 errors.Is(err, New(code)) 按错误码匹配。
func (e *BizError) Is(target error) bool {
	t, ok := target.(*BizError)
	if !ok {
		return false
	}
	return e.Code == t.Code
}

// IsBizError 断言 err 是否为 *BizError，同时返回类型断言结果。
func IsBizError(err error) (*BizError, bool) {
	var be *BizError
	if errors.As(err, &be) {
		return be, true
	}
	return nil, false
}
