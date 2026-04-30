package handler

import (
	"ai_interview/internal/service"
	"net/http"
)

// Router 聚合所有子模块 handler，统一注册路由。
type Router struct {
	auth          *authHandler
	device        *deviceHandler
	resume        *resumeHandler
	interview     *interviewHandler
	report        *reportHandler
	questionnaire *questionnaireHandler
}

func NewRouter(interviewSvc service.InterviewService) *Router {
	return &Router{
		auth:          &authHandler{},
		device:        &deviceHandler{},
		resume:        &resumeHandler{},
		interview:     NewInterviewHandler(interviewSvc),
		report:        &reportHandler{},
		questionnaire: &questionnaireHandler{},
	}
}

// Register 按 API 文档注册所有路由，路径统一加 /v1 前缀。
func (r *Router) Register(mux *http.ServeMux) {
	// 认证模块
	mux.HandleFunc("POST /v1/auth/register", r.auth.Register)
	mux.HandleFunc("POST /v1/auth/login", r.auth.Login)
	mux.HandleFunc("POST /v1/auth/guest", r.auth.Guest)

	// 设备检测模块
	mux.HandleFunc("POST /v1/device/check", r.device.Check)

	// 简历模块
	mux.HandleFunc("POST /v1/resume/parse", r.resume.Parse)
	mux.HandleFunc("POST /v1/resume/submit", r.resume.Submit)

	// 面试配置模块
	mux.HandleFunc("POST /v1/interview/config", r.interview.Config)

	// 面试模块
	mux.HandleFunc("POST /v1/interview/create", r.interview.Create)
	mux.HandleFunc("GET /v1/interview/stream", r.interview.Stream)
	mux.HandleFunc("POST /v1/interview/audio", r.interview.Audio)
	mux.HandleFunc("POST /v1/interview/finish", r.interview.Finish)
	mux.HandleFunc("GET /v1/interview/state", r.interview.State)

	// 代码提交模块
	mux.HandleFunc("POST /v1/interview/code/submit", r.interview.CodeSubmit)

	// 报告模块
	mux.HandleFunc("GET /v1/report/status", r.report.Status)
	mux.HandleFunc("GET /v1/report", r.report.Get)

	// 问卷模块
	mux.HandleFunc("GET /v1/questionnaire", r.questionnaire.Get)
	mux.HandleFunc("POST /v1/questionnaire/submit", r.questionnaire.Submit)
}
