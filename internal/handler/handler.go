package handler

import (
	"context"
	"time"

	"github.com/cloudwego/hertz/pkg/app/server"
	"github.com/redis/go-redis/v9"

	"ai_interview/internal/config"
	"ai_interview/internal/log"
	"ai_interview/internal/middleware"
	"ai_interview/internal/middleware/ratelimit"
	"ai_interview/internal/router"
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
	// Rdb 用于限流中间件（Lua 滑窗脚本），由 app 层注入
	Rdb *redis.Client
}

// Server 封装 HTTP 服务生命周期，对外隐藏 Hertz 实现细节。
type Server struct {
	h *server.Hertz
}

// NewServer 创建并配置好服务实例，注册所有路由。
func NewServer(cfg *config.Config, svc Services) *Server {
	h := server.Default(server.WithHostPorts(cfg.HTTPAddr))
	h.Use(middleware.Logger()) // 全局：注入 request_id + 访问日志

	wsLimiter := ratelimit.NewLimiter(svc.Rdb, map[ratelimit.Dimension]ratelimit.Config{
		ratelimit.DimensionIP:   {Limit: 30, Window: time.Minute},
		ratelimit.DimensionUser: {Limit: 60, Window: time.Minute},
	})

	router.Register(h, router.Deps{
		JWTSecret: cfg.JWTSecret,
		Rdb:       svc.Rdb,
		Auth:      &authHandler{svc: svc.Auth},
		Device:    &deviceHandler{svc: svc.Device},
		Resume:    &resumeHandler{svc: svc.Resume},
		Interview: &interviewHandler{svc: svc.Interview},
		WSInterview: &wsInterviewHandler{
			jwtSecret: cfg.JWTSecret,
			limiter:   wsLimiter,
		},
		Report:        &reportHandler{svc: svc.Report},
		Questionnaire: &questionnaireHandler{svc: svc.Questionnaire},
	})
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
