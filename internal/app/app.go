package app

import (
	"context"
	"fmt"

	"ai_interview/internal/config"
	"ai_interview/internal/einocore/agent"
	"ai_interview/internal/einocore/compose"
	"ai_interview/internal/handler"
	"ai_interview/internal/service"
	"ai_interview/internal/storage/postgres"
	sredis "ai_interview/internal/storage/redis"
	"ai_interview/internal/storage/s3"
)

// App 持有所有依赖实例，按顺序初始化。
type App struct {
	Server *handler.Server
	db     *postgres.DB
	redis  *sredis.Client
	s3     *s3.Client
}

// New 按依赖顺序初始化所有组件，返回可运行的 App 实例。
func New(cfg *config.Config) (*App, error) {
	ctx := context.Background()

	// 1. PostgreSQL
	db, err := postgres.New(ctx, cfg.PostgresDSN)
	if err != nil {
		return nil, fmt.Errorf("init postgres: %w", err)
	}
	if err := postgres.RunMigrations(ctx, db.Conn(), "migrations"); err != nil {
		return nil, fmt.Errorf("run migrations: %w", err)
	}

	// 2. Redis
	rdb, err := sredis.New(ctx, sredis.Options{
		Addr:     cfg.RedisAddr,
		Password: cfg.RedisPassword,
		DB:       cfg.RedisDB,
	})
	if err != nil {
		return nil, fmt.Errorf("init redis: %w", err)
	}

	// 3. S3 / OSS
	s3Client, err := s3.New(ctx, s3.Options{
		Endpoint:  cfg.S3Endpoint,
		AccessKey: cfg.S3AccessKey,
		SecretKey: cfg.S3SecretKey,
		Bucket:    cfg.S3Bucket,
		Region:    cfg.S3Region,
		UseSSL:    cfg.S3UseSSL,
	})
	if err != nil {
		return nil, fmt.Errorf("init s3: %w", err)
	}

	// 4. Session 管理
	sessionManager := service.NewSessionManager(rdb.Client(), cfg.InterviewStateTTL)

	// 5. AI 层
	supervisor, err := agent.NewSupervisor()
	if err != nil {
		return nil, fmt.Errorf("new supervisor: %w", err)
	}
	graph, err := compose.NewInterviewGraph(ctx, supervisor)
	if err != nil {
		return nil, fmt.Errorf("new interview graph: %w", err)
	}

	// 6. Service 层
	interviewSvc := service.NewInterviewService(sessionManager, graph)

	// 7. HTTP Server
	srv := handler.NewServer(cfg, handler.Services{
		Interview: interviewSvc,
	})

	return &App{Server: srv, db: db, redis: rdb, s3: s3Client}, nil
}

// Run 启动服务，阻塞直到服务退出。
func (a *App) Run() {
	a.Server.Run()
}

// Shutdown 优雅关闭所有服务。
func (a *App) Shutdown() {
	a.Server.Shutdown()
	a.redis.Close()
	a.db.Close()
}
