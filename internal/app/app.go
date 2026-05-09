package app

import (
	"context"
	"fmt"

	"ai_interview/internal/auth"
	"ai_interview/internal/config"
	"ai_interview/internal/einocore/agent"
	"ai_interview/internal/einocore/compose"
	"ai_interview/internal/handler"
	"ai_interview/internal/llm"
	"ai_interview/internal/service"
	"ai_interview/internal/storage/es"
	"ai_interview/internal/storage/milvus"
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
	milvus *milvus.Client
	es     *es.Client
}

// New 按依赖顺序初始化所有组件，返回可运行的 App 实例。
func New(cfg *config.Config) (*App, error) {
	ctx := context.Background()

	// 1. PostgreSQL
	db, err := postgres.New(ctx, postgres.Options{
		DSN:             cfg.PostgresDSN,
		MaxOpenConns:    cfg.PGMaxOpenConns,
		MaxIdleConns:    cfg.PGMaxIdleConns,
		ConnMaxLifetime: cfg.PGConnMaxLifetime,
		ConnMaxIdleTime: cfg.PGConnMaxIdleTime,
	})
	if err != nil {
		return nil, fmt.Errorf("init postgres: %w", err)
	}
	if err := postgres.RunMigrations(ctx, db.Conn(), "migrations"); err != nil {
		return nil, fmt.Errorf("run migrations: %w", err)
	}

	// 2. Redis
	rdb, err := sredis.New(ctx, sredis.Options{
		Addr:         cfg.RedisAddr,
		Password:     cfg.RedisPassword,
		DB:           cfg.RedisDB,
		PoolSize:     cfg.RedisPoolSize,
		MinIdleConns: cfg.RedisMinIdleConns,
		DialTimeout:  cfg.RedisDialTimeout,
		ReadTimeout:  cfg.RedisReadTimeout,
		WriteTimeout: cfg.RedisWriteTimeout,
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

	// 4. Milvus（向量数据库）
	milvusClient, err := milvus.New(ctx, milvus.Options{
		Addr:       cfg.MilvusAddr,
		Collection: cfg.MilvusCollection,
	})
	if err != nil {
		return nil, fmt.Errorf("init milvus: %w", err)
	}
	if err := milvusClient.EnsureCollection(ctx); err != nil {
		return nil, fmt.Errorf("milvus ensure collection: %w", err)
	}

	// 5. Elasticsearch（关键词/标签检索）
	esClient, err := es.New(ctx, es.Options{
		Addrs:    cfg.ESAddrs,
		Username: cfg.ESUsername,
		Password: cfg.ESPassword,
		Index:    cfg.ESIndex,
	})
	if err != nil {
		return nil, fmt.Errorf("init es: %w", err)
	}
	if err := esClient.EnsureIndex(ctx); err != nil {
		return nil, fmt.Errorf("es ensure index: %w", err)
	}

	// 6. Session 管理
	llm.Init(cfg) // LLM provider registry 初始化（静态配置，无 IO）
	sessionManager := service.NewSessionManager(rdb.Client(), cfg.InterviewStateTTL)

	// 7. AI 层
	supervisor, err := agent.NewSupervisor()
	if err != nil {
		return nil, fmt.Errorf("new supervisor: %w", err)
	}
	graph, err := compose.NewInterviewGraph(ctx, supervisor)
	if err != nil {
		return nil, fmt.Errorf("new interview graph: %w", err)
	}

	// 8. Service 层
	userRepo := postgres.NewUserRepo(db.Conn())
	jwtCfg := auth.TokenConfig{
		Secret:    cfg.JWTSecret,
		Issuer:    cfg.JWTIssuer,
		ExpMinute: cfg.JWTAccessExpMin,
	}
	authSvc := service.NewAuthService(userRepo, jwtCfg)
	resumeSvc := service.NewResumeService(s3Client)
	interviewSvc := service.NewInterviewService(sessionManager, graph)

	// 9. HTTP Server
	srv := handler.NewServer(cfg, handler.Services{
		Auth:      authSvc,
		Resume:    resumeSvc,
		Interview: interviewSvc,
	})

	return &App{Server: srv, db: db, redis: rdb, s3: s3Client, milvus: milvusClient, es: esClient}, nil
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
	a.milvus.Close()
}
