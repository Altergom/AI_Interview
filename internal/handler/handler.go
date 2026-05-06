package handler

import (
	"context"
	"net/http"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/app/server"

	"ai_interview/internal/config"
	"ai_interview/internal/infra/log"
	"ai_interview/internal/service"
)

// Services 聚合所有业务服务，由外部注入。
type Services struct {
	Interview     service.InterviewService
	Auth          service.AuthService
	Resume        service.ResumeService
	Device        service.DeviceService
	Report        service.ReportService
	Questionnaire service.QuestionnaireService
}

// Server 封装 HTTP 服务生命周期，对外隐藏 Hertz 实现细节。
type Server struct {
	h *server.Hertz
}

// NewServer 创建并配置好服务实例，注册所有路由。
func NewServer(cfg *config.Config, svc Services) *Server {
	h := server.Default(server.WithHostPorts(cfg.HTTPAddr))
	newRouter(svc).register(h)
	return &Server{h: h}
}

// Run 启动服务，阻塞直到服务退出。
func (s *Server) Run() {
	s.h.Spin()
}

// Shutdown 优雅关闭服务。
func (s *Server) Shutdown() {
	log.Info("shutting down")
	s.h.Shutdown(context.Background())
}

// Router 聚合所有子模块 handler，统一注册路由。
type Router struct {
	auth          *authHandler
	device        *deviceHandler
	resume        *resumeHandler
	interview     *interviewHandler
	report        *reportHandler
	questionnaire *questionnaireHandler
}

func newRouter(svc Services) *Router {
	return &Router{
		auth:          &authHandler{svc: svc.Auth},
		device:        &deviceHandler{svc: svc.Device},
		resume:        &resumeHandler{svc: svc.Resume},
		interview:     &interviewHandler{svc: svc.Interview},
		report:        &reportHandler{svc: svc.Report},
		questionnaire: &questionnaireHandler{svc: svc.Questionnaire},
	}
}

// register 按 API 文档注册所有路由，路径统一加 /v1 前缀。
func (r *Router) register(h *server.Hertz) {
	h.GET("/", func(ctx context.Context, c *app.RequestContext) {
		c.JSON(http.StatusOK, map[string]string{
			"service": "ai_interview",
			"role":    "api",
		})
	})

	// 存活探针
	h.GET("/health", func(ctx context.Context, c *app.RequestContext) {
		c.JSON(http.StatusOK, map[string]string{"status": "ok"})
	})
	h.GET("/healthz", func(ctx context.Context, c *app.RequestContext) {
		c.JSON(http.StatusOK, map[string]string{"status": "ok"})
	})

	v1 := h.Group("/v1")

	// 认证模块
	auth := v1.Group("/auth")
	auth.POST("/register", r.auth.Register)
	auth.POST("/login", r.auth.Login)
	auth.POST("/guest", r.auth.Guest)

	// 设备检测模块
	v1.POST("/device/check", r.device.Check)

	// 简历模块
	resume := v1.Group("/resume")
	resume.POST("/parse", r.resume.Parse)
	resume.POST("/submit", r.resume.Submit)

	// 面试模块
	interview := v1.Group("/interview")
	interview.POST("/config", r.interview.Config)
	interview.POST("/create", r.interview.Create)
	interview.GET("/stream", r.interview.Stream)
	interview.POST("/audio", r.interview.Audio)
	interview.POST("/finish", r.interview.Finish)
	interview.GET("/state", r.interview.State)
	interview.POST("/code/submit", r.interview.CodeSubmit)

	// 报告模块
	report := v1.Group("/report")
	report.GET("/status", r.report.Status)
	report.GET("", r.report.Get)

	// 问卷模块
	questionnaire := v1.Group("/questionnaire")
	questionnaire.GET("", r.questionnaire.Get)
	questionnaire.POST("/submit", r.questionnaire.Submit)
}
