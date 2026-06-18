package handler

import "net/http"

// Result 统一响应格式，与主服务保持一致。
type Result struct {
	Success bool   `json:"success"`
	Data    any    `json:"data"`
	Error   *Error `json:"error"`
}

// Error 错误详情。
type Error struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

func ok(data any) (int, *Result) {
	return http.StatusOK, &Result{Success: true, Data: data}
}

func fail(status, code int, msg string) (int, *Result) {
	return status, &Result{
		Success: false,
		Data:    nil,
		Error:   &Error{Code: code, Message: msg},
	}
}
