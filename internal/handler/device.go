package handler

import (
	"context"
	"net/http"

	"github.com/cloudwego/hertz/pkg/app"

	"ai_interview/internal/service"
	biz "ai_interview/internal/utils/respx"
)

type deviceHandler struct {
	svc service.DeviceService
}

type deviceCheckReq struct {
	HasMicrophone bool   `json:"has_microphone"`
	HasCamera     bool   `json:"has_camera"`
	Browser       string `json:"browser"`
	OS            string `json:"os"`
}

// Check POST /v1/device/check
func (h *deviceHandler) Check(ctx context.Context, c *app.RequestContext) {
	var req deviceCheckReq
	if err := c.BindJSON(&req); err != nil {
		Fail(ctx, c, http.StatusBadRequest, biz.CodeBadRequest, "invalid request body")
		return
	}

	if h.svc == nil {
		OK(ctx, c, map[string]string{
			"status":  "ok",
			"message": "device info recorded",
		})
		return
	}

	result, err := h.svc.Check(ctx, service.DeviceCheckRequest{
		HasMicrophone: req.HasMicrophone,
		HasCamera:     req.HasCamera,
		Browser:       req.Browser,
		OS:            req.OS,
	})
	if err != nil {
		HandleErr(ctx, c, err)
		return
	}

	OK(ctx, c, result)
}
