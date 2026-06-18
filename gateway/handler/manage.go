package handler

import (
	"context"
	"net/http"

	"github.com/cloudwego/hertz/pkg/app"

	"gateway/domain"
	"gateway/session"
)

// ManageHandler 处理管理操作接口。
type ManageHandler struct {
	mgr *session.Manager
}

func NewManageHandler(mgr *session.Manager) *ManageHandler {
	return &ManageHandler{mgr: mgr}
}

type handoffReq struct {
	Reason     string `json:"reason"`
	OperatorID string `json:"operator_id"`
}

// Handoff POST /v1/gateway/session/:session_id/handoff
func (h *ManageHandler) Handoff(ctx context.Context, c *app.RequestContext) {
	sessionID := c.Param("session_id")

	var req handoffReq
	if err := c.BindJSON(&req); err != nil {
		status, resp := fail(http.StatusBadRequest, 400, "invalid request body")
		c.JSON(status, resp)
		return
	}

	if err := h.mgr.UpdateStatus(ctx, sessionID, domain.StatusHandoff); err != nil {
		status, resp := fail(http.StatusInternalServerError, 500, "handoff failed")
		c.JSON(status, resp)
		return
	}

	status, resp := ok(map[string]any{
		"session_id": sessionID,
		"status":     domain.StatusHandoff,
	})
	c.JSON(status, resp)
}

type terminateReq struct {
	Reason     string `json:"reason"`
	OperatorID string `json:"operator_id"`
}

// Terminate POST /v1/gateway/session/:session_id/terminate
func (h *ManageHandler) Terminate(ctx context.Context, c *app.RequestContext) {
	sessionID := c.Param("session_id")

	var req terminateReq
	if err := c.BindJSON(&req); err != nil {
		status, resp := fail(http.StatusBadRequest, 400, "invalid request body")
		c.JSON(status, resp)
		return
	}

	if err := h.mgr.UpdateStatus(ctx, sessionID, domain.StatusFinished); err != nil {
		status, resp := fail(http.StatusInternalServerError, 500, "terminate failed")
		c.JSON(status, resp)
		return
	}

	status, resp := ok(map[string]any{
		"session_id": sessionID,
		"status":     domain.StatusFinished,
	})
	c.JSON(status, resp)
}
