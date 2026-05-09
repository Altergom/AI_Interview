package handler

import (
	"context"
	"net/http"
	"time"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/app/server"
	"github.com/redis/go-redis/v9"

	"ai_interview/internal/config"
	"ai_interview/internal/log"
	"ai_interview/internal/middleware"
	authmw "ai_interview/internal/middleware/auth"
	"ai_interview/internal/middleware/ratelimit"
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
	newRouter(cfg.JWTSecret, svc).register(h)
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
	jwtSecret     string
	rdb           *redis.Client // 限流用
	auth          *authHandler
	device        *deviceHandler
	resume        *resumeHandler
	interview     *interviewHandler
	wsInterview   *wsInterviewHandler
	report        *reportHandler
	questionnaire *questionnaireHandler
}

func newRouter(jwtSecret string, svc Services) *Router {
	// WebSocket 连接建立限流：相比普通 HTTP 接口更宽松
	// （连接建立只是握手，实际流量由消息体积决定）
	wsLimiter := ratelimit.NewLimiter(svc.Rdb, map[ratelimit.Dimension]ratelimit.Config{
		ratelimit.DimensionIP:   {Limit: 30, Window: time.Minute},
		ratelimit.DimensionUser: {Limit: 60, Window: time.Minute},
	})

	return &Router{
		jwtSecret:     jwtSecret,
		rdb:           svc.Rdb,
		auth:          &authHandler{svc: svc.Auth},
		device:        &deviceHandler{svc: svc.Device},
		resume:        &resumeHandler{svc: svc.Resume},
		interview:     &interviewHandler{svc: svc.Interview},
		wsInterview:   &wsInterviewHandler{jwtSecret: jwtSecret, limiter: wsLimiter},
		report:        &reportHandler{svc: svc.Report},
		questionnaire: &questionnaireHandler{svc: svc.Questionnaire},
	}
}

// register 按 API 文档注册所有路由。
//
// 路由分层：
//   - 公开路由：健康检查、auth 三个端点、设备检测
//   - 受保护路由（HAuth）：简历、面试、报告、问卷
//   - WS 路由：鉴权在 handler 内握手阶段完成（不走 HAuth 中间件）
func (r *Router) register(h *server.Hertz) {
	// ── 公开端点 ─────────────────────────────────────────────────────────────
	h.GET("/", func(ctx context.Context, c *app.RequestContext) {
		c.JSON(http.StatusOK, map[string]string{
			"service": "ai_interview",
			"role":    "api",
		})
	})
	h.GET("/health", func(ctx context.Context, c *app.RequestContext) {
		c.JSON(http.StatusOK, map[string]string{"status": "ok"})
	})
	h.GET("/healthz", func(ctx context.Context, c *app.RequestContext) {
		c.JSON(http.StatusOK, map[string]string{"status": "ok"})
	})

	v1 := h.Group("/v1")

	// 认证模块（无需鉴权）
	auth := v1.Group("/auth")
	auth.POST("/register", r.auth.Register)
	auth.POST("/login", r.auth.Login)
	auth.POST("/guest", r.auth.Guest)

	// 设备检测（无需鉴权，麦克风测试在登录前进行）
	v1.POST("/device/check", r.device.Check)

	// ── 受保护端点（HAuth 校验 JWT）────────────────────────────────────────
	hauth := authmw.HAuth(r.jwtSecret)

	// 简历模块
	resume := v1.Group("/resume", hauth)
	resume.GET("/upload-url", r.resume.PresignUpload)
	// /parse 接 IP+USER 双维度限流（10/min IP, 30/min USER）
	resume.POST("/parse", ratelimit.Middleware(r.rdb, "resume.parse"), r.resume.Parse)
	resume.POST("/submit", r.resume.Submit)
	resume.GET("", r.resume.Get)

	// 面试模块（HTTP）
	interview := v1.Group("/interview", hauth)
	interview.POST("/config", r.interview.Config)
	interview.POST("/create", r.interview.Create)
	interview.GET("/stream", r.interview.Stream)
	interview.POST("/audio", r.interview.Audio)
	interview.POST("/finish", r.interview.Finish)
	interview.GET("/state", r.interview.State)
	interview.POST("/code/submit", r.interview.CodeSubmit)

	// ── WebSocket 端点（鉴权在 handler 内握手阶段完成）────────────────────
	// 注意：不挂 HAuth 中间件，因为浏览器 WebSocket API 不支持自定义 header，
	// JWT 改为 query param 传入，在 ServeWS 内部验证。
	v1.GET("/interview/ws/:interview_id", r.wsInterview.ServeWS)

	// 报告模块
	report := v1.Group("/report", hauth)
	report.GET("/status", r.report.Status)
	report.GET("", r.report.Get)

	// 问卷模块
	questionnaire := v1.Group("/questionnaire", hauth)
	questionnaire.GET("", r.questionnaire.Get)
	questionnaire.POST("/submit", r.questionnaire.Submit)
}
