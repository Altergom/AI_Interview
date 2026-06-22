package handler

import (
	"context"
	"io"
	"net/http"

	"github.com/cloudwego/hertz/pkg/app"

	authmw "ai_interview/internal/middleware/auth"
	"ai_interview/internal/service"
	"ai_interview/internal/utils/uuidx"
	biz "ai_interview/internal/utils/respx"
)

type interviewHandler struct {
	svc service.InterviewService
}

type configReq struct {
	Position  string `json:"position"`
	Direction string `json:"direction"`
}

// Config POST /v1/interview/config
func (h *interviewHandler) Config(ctx context.Context, c *app.RequestContext) {
	userID := authmw.GetUserID(c)

	var req configReq
	if err := c.BindJSON(&req); err != nil {
		Fail(ctx, c, http.StatusBadRequest, biz.CodeBadRequest, "invalid request body")
		return
	}
	if req.Direction == "" {
		Fail(ctx, c, http.StatusBadRequest, biz.CodeBadRequest, "direction is required")
		return
	}

	interviewID, err := h.svc.SetConfig(ctx, service.InterviewConfigRequest{
		UserID:    userID,
		Position:  req.Position,
		Direction: req.Direction,
	})
	if err != nil {
		HandleErr(ctx, c, err)
		return
	}

	OK(ctx, c, map[string]string{"interview_id": interviewID})
}

type createReq struct {
	InterviewID string `json:"interview_id"`
}

// Create POST /v1/interview/create
func (h *interviewHandler) Create(ctx context.Context, c *app.RequestContext) {
	userID := authmw.GetUserID(c)

	var req createReq
	if err := c.BindJSON(&req); err != nil {
		Fail(ctx, c, http.StatusBadRequest, biz.CodeBadRequest, "invalid request body")
		return
	}
	if req.InterviewID == "" {
		Fail(ctx, c, http.StatusBadRequest, biz.CodeBadRequest, "interview_id is required")
		return
	}

	result, err := h.svc.Create(ctx, req.InterviewID, userID)
	if err != nil {
		HandleErr(ctx, c, err)
		return
	}

	OK(ctx, c, result)
}

// Stream GET /v1/interview/stream?interview_id={}
func (h *interviewHandler) Stream(ctx context.Context, c *app.RequestContext) {
	interviewID := string(c.Query("interview_id"))
	if interviewID == "" {
		Fail(ctx, c, http.StatusBadRequest, biz.CodeBadRequest, "interview_id is required")
		return
	}

	state, err := h.svc.GetState(ctx, interviewID)
	if err != nil {
		HandleErr(ctx, c, err)
		return
	}

	OK(ctx, c, state)
}

// Audio POST /v1/interview/audio
func (h *interviewHandler) Audio(ctx context.Context, c *app.RequestContext) {
	interviewID := string(c.GetHeader("X-Interview-Id"))
	turnID := string(c.GetHeader("X-Turn-Id"))

	if interviewID == "" {
		Fail(ctx, c, http.StatusBadRequest, biz.CodeBadRequest, "X-Interview-Id header is required")
		return
	}
	if turnID == "" {
		turnID = uuidx.NewShort(8)
	}

	body, err := io.ReadAll(c.RequestBodyStream())
	if err != nil {
		Fail(ctx, c, http.StatusBadRequest, biz.CodeBadRequest, "failed to read audio body")
		return
	}
	if len(body) == 0 {
		Fail(ctx, c, http.StatusBadRequest, biz.CodeBadRequest, "empty audio body")
		return
	}

	if err := h.svc.ProcessAudio(ctx, service.AudioRequest{
		InterviewID: interviewID,
		TurnID:      turnID,
		AudioData:   body,
	}); err != nil {
		HandleErr(ctx, c, err)
		return
	}

	OK(ctx, c, map[string]string{"status": "accepted", "turn_id": turnID})
}

// Finish POST /v1/interview/finish
func (h *interviewHandler) Finish(ctx context.Context, c *app.RequestContext) {
	type finishReq struct {
		InterviewID string `json:"interview_id"`
	}

	var req finishReq
	if err := c.BindJSON(&req); err != nil {
		Fail(ctx, c, http.StatusBadRequest, biz.CodeBadRequest, "invalid request body")
		return
	}
	if req.InterviewID == "" {
		Fail(ctx, c, http.StatusBadRequest, biz.CodeBadRequest, "interview_id is required")
		return
	}

	result, err := h.svc.Finish(ctx, req.InterviewID)
	if err != nil {
		HandleErr(ctx, c, err)
		return
	}

	OK(ctx, c, result)
}

// State GET /v1/interview/state?interview_id={}
func (h *interviewHandler) State(ctx context.Context, c *app.RequestContext) {
	interviewID := string(c.Query("interview_id"))
	if interviewID == "" {
		Fail(ctx, c, http.StatusBadRequest, biz.CodeBadRequest, "interview_id is required")
		return
	}

	state, err := h.svc.GetState(ctx, interviewID)
	if err != nil {
		HandleErr(ctx, c, err)
		return
	}

	OK(ctx, c, state)
}

// CodeSubmit POST /v1/interview/code/submit
func (h *interviewHandler) CodeSubmit(ctx context.Context, c *app.RequestContext) {
	type codeSubmitReq struct {
		InterviewID string `json:"interview_id"`
		QuestionID  string `json:"question_id"`
		Language    string `json:"language"`
		Code        string `json:"code"`
	}

	var req codeSubmitReq
	if err := c.BindJSON(&req); err != nil {
		Fail(ctx, c, http.StatusBadRequest, biz.CodeBadRequest, "invalid request body")
		return
	}
	if req.InterviewID == "" || req.Code == "" {
		Fail(ctx, c, http.StatusBadRequest, biz.CodeBadRequest, "interview_id and code are required")
		return
	}

	if err := h.svc.SubmitCode(ctx, service.CodeSubmitRequest{
		InterviewID: req.InterviewID,
		QuestionID:  req.QuestionID,
		Language:    req.Language,
		Code:        req.Code,
	}); err != nil {
		HandleErr(ctx, c, err)
		return
	}

	OK(ctx, c, map[string]string{"status": "accepted"})
}
