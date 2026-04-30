package handler

import (
	"ai_interview/internal/service"
	"encoding/json"
	"net/http"
	"strings"
)

const codeInvalidParam = 1001

type interviewHandler struct {
	svc service.InterviewService
}

func NewInterviewHandler(svc service.InterviewService) *interviewHandler {
	return &interviewHandler{
		svc: svc,
	}
}

type interviewConfigReq struct {
	UserID    string `json:"user_id"`
	Position  string `json:"position"`
	Direction string `json:"direction"`
}

func validPosition(position string) bool {
	switch position {
	case "golang", "java", "frontend", "test":
		return true
	default:
		return false
	}
}

func validDirection(direction string) bool {
	switch direction {
	case "backend", "cloud", "agent", "server":
		return true
	default:
		return false
	}
}

// Config POST /v1/interview/config
// 设置面试岗位（position）和方向（direction），配置存入 Redis。
func (h *interviewHandler) Config(w http.ResponseWriter, r *http.Request) {
	var req interviewConfigReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, Resp{Code: codeInvalidParam, Data: "invalid json"})
		return
	}

	req.UserID = strings.TrimSpace(req.UserID)
	req.Position = strings.ToLower(strings.TrimSpace(req.Position))
	req.Direction = strings.ToLower(strings.TrimSpace(req.Direction))

	if req.UserID == "" || req.Position == "" || req.Direction == "" {
		writeJSON(w, http.StatusBadRequest, Resp{Code: codeInvalidParam, Data: "user_id, position, direction are required"})
		return
	}

	if !validPosition(req.Position) {
		writeJSON(w, http.StatusBadRequest, Resp{Code: codeInvalidParam, Data: "invalid position"})
		return
	}
	if !validDirection(req.Direction) {
		writeJSON(w, http.StatusBadRequest, Resp{Code: codeInvalidParam, Data: "invalid direction"})
		return
	}
	if h.svc == nil {
		writeJSON(w, http.StatusInternalServerError, Resp{Code: 500, Data: "interview service not configured"})
		return
	}

	configID, err := h.svc.SetConfig(r.Context(), service.InterviewConfigRequest{
		UserID:    req.UserID,
		Position:  req.Position,
		Direction: req.Direction,
	})
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, Resp{Code: 500, Data: err.Error()})
		return
	}

	ok(w, map[string]string{
		"config_id": configID,
		"message":   "配置保存成功",
	})

}

// Create POST /v1/interview/create
// 创建面试会话，返回 interview_id 及初始阶段 intro。
func (h *interviewHandler) Create(w http.ResponseWriter, r *http.Request) {
	panic("not implemented")
}

// Stream GET /v1/interview/stream?interview_id={}
// 建立 SSE 连接，服务端推送 AI 文字流、音频流及阶段事件。
func (h *interviewHandler) Stream(w http.ResponseWriter, r *http.Request) {
	panic("not implemented")
}

// Audio POST /v1/interview/audio  (application/octet-stream)
// Headers: X-Interview-Id, X-Turn-Id
// 接收前端 VAD 截断后的原始音频，触发 ASR → Router → Interview Agent 链路。
func (h *interviewHandler) Audio(w http.ResponseWriter, r *http.Request) {
	panic("not implemented")
}

// Finish POST /v1/interview/finish
// 结束面试，发布 interview_finished 消息至 MQ，异步生成报告。
func (h *interviewHandler) Finish(w http.ResponseWriter, r *http.Request) {
	panic("not implemented")
}

// State GET /v1/interview/state?interview_id={}
// 查询面试当前阶段、已提问题数等状态。
func (h *interviewHandler) State(w http.ResponseWriter, r *http.Request) {
	panic("not implemented")
}

// CodeSubmit POST /v1/interview/code/submit
// 提交算法题代码，触发 Code Judge Agent → Interview Agent 链路。
func (h *interviewHandler) CodeSubmit(w http.ResponseWriter, r *http.Request) {
	panic("not implemented")
}
