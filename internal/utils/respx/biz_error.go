package respx

import (
	"errors"
	"fmt"
)

type BizError struct {
	Code    ErrorCode
	Message string
	Cause   error
}

func New(code ErrorCode) *BizError {
	return &BizError{Code: code, Message: code.Message()}
}

func NewMsg(code ErrorCode, msg string) *BizError {
	return &BizError{Code: code, Message: msg}
}

func Wrap(code ErrorCode, cause error) *BizError {
	return &BizError{Code: code, Message: code.Message(), Cause: cause}
}

func WrapMsg(code ErrorCode, msg string, cause error) *BizError {
	return &BizError{Code: code, Message: msg, Cause: cause}
}

func (e *BizError) Error() string {
	if e.Cause != nil {
		return fmt.Sprintf("biz_error[%d]: %s: %v", e.Code, e.Message, e.Cause)
	}
	return fmt.Sprintf("biz_error[%d]: %s", e.Code, e.Message)
}

func (e *BizError) Unwrap() error {
	return e.Cause
}

func (e *BizError) Is(target error) bool {
	t, ok := target.(*BizError)
	if !ok {
		return false
	}
	return e.Code == t.Code
}

func IsBizError(err error) (*BizError, bool) {
	var be *BizError
	if errors.As(err, &be) {
		return be, true
	}
	return nil, false
}

