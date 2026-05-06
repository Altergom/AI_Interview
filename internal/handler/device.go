package handler

import (
	"context"

	"github.com/cloudwego/hertz/pkg/app"

	"ai_interview/internal/service"
)

type deviceHandler struct {
	svc service.DeviceService
}

// Check POST /v1/device/check
// 前端完成权限获取后调用，服务端记录设备信息。
func (h *deviceHandler) Check(ctx context.Context, c *app.RequestContext) {
	panic("not implemented")
}
