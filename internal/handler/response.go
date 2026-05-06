package handler

import (
	"context"
	"net/http"

	"github.com/cloudwego/hertz/pkg/app"
)

// 错误码，与 API 文档「错误码」一节保持一致。
const (
	CodeOK = 0

	// 通用
	CodeBadRequest   = 1001
	CodeUnauthorized = 1002

	// 用户
	CodeUserNotFound    = 2001
	CodeEmailRegistered = 2002

	// 简历
	CodeResumeParseFailed = 3001
	CodeResumeFormatUnsup = 3002

	// 面试
	CodeInterviewNotFound   = 4001
	CodeInterviewFinished   = 4002
	CodeInvalidStageForCode = 4003

	// AI
	CodeASRFailed = 5001
	CodeLLMFailed = 5002
	CodeTTSFailed = 5003

	// 报告
	CodeReportGenerating = 6001
	CodeReportNotFound   = 6002
)

// Resp 统一响应格式
type Resp struct {
	Code int    `json:"code"`
	Msg  string `json:"msg,omitempty"`
	Data any    `json:"data,omitempty"`
}

func ok(ctx context.Context, c *app.RequestContext, data any) {
	c.JSON(http.StatusOK, Resp{Code: CodeOK, Data: data})
}

func fail(ctx context.Context, c *app.RequestContext, httpStatus int, code int, msg string) {
	c.JSON(httpStatus, Resp{Code: code, Msg: msg})
}
