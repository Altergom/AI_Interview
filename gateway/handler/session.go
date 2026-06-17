package handler

import (
	"context"
	"net/http"
	"strconv"

	"github.com/cloudwego/hertz/pkg/app"

	"gateway/session"
)

// SessionHandler 处理会话查询接口。
type SessionHandler struct {
	mgr *session.Manager
}

func NewSessionHandler(mgr *session.Manager) *SessionHandler {
	return &SessionHandler{mgr: mgr}
}

// Get GET /v1/gateway/session/:session_id
func (h *SessionHandler) Get(ctx context.Context, c *app.RequestContext) {
	sessionID := c.Param("session_id")
	if sessionID == "" {
		status, resp := fail(http.StatusBadRequest, 400, "session_id is required")
		c.JSON(status, resp)
		return
	}

	s, err := h.mgr.Get(ctx, sessionID)
	if err != nil {
		status, resp := fail(http.StatusNotFound, 404, "session not found")
		c.JSON(status, resp)
		return
	}

	status, resp := ok(s)
	c.JSON(status, resp)
}

// List GET /v1/gateway/sessions?candidate_id=&page=&page_size=
func (h *SessionHandler) List(ctx context.Context, c *app.RequestContext) {
	candidateID := string(c.Query("candidate_id"))
	page, _ := strconv.Atoi(string(c.DefaultQuery("page", "1")))
	pageSize, _ := strconv.Atoi(string(c.DefaultQuery("page_size", "20")))

	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	sessions, total, err := h.mgr.List(ctx, candidateID, page, pageSize)
	if err != nil {
		status, resp := fail(http.StatusInternalServerError, 500, "list sessions failed")
		c.JSON(status, resp)
		return
	}

	status, resp := ok(map[string]any{
		"sessions":  sessions,
		"total":     total,
		"page":      page,
		"page_size": pageSize,
	})
	c.JSON(status, resp)
}
