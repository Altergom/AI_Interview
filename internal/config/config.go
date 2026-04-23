package config

import "time"

// App 从环境变量与 .env 文件加载的集中配置（与 技术文档 / TODO「基础设施」一致）。
type App struct {
	Env       string
	LogLevel  string
	LogFormat string // json | text；json 便于接入 Loki / ELK / 云厂商日志采集

	// HTTP（网关 / 健康检查等）
	HTTPAddr string

	// PostgreSQL
	PostgresDSN string

	// Redis
	RedisAddr     string
	RedisPassword string
	RedisDB       int

	// 消息队列
	MQBrokerURL              string
	MQTopicInterviewFinished string

	// 对象存储（S3 兼容 API）
	S3Endpoint  string
	S3AccessKey string
	S3SecretKey string
	S3Bucket    string
	S3Region    string
	S3UseSSL    bool

	// JWT
	JWTSecret       string
	JWTIssuer       string
	JWTAccessExpMin int

	// LLM / Eino（按后续接入的 provider 选用）
	OpenAIAPIKey  string
	OpenAIBaseURL string
	DoubaoAPIKey  string
	DoubaoBaseURL string

	// TTL：可由 RESUME_REDIS_TTL、INTERVIEW_STATE_TTL 等 duration 字符串覆盖
	ResumeRedisTTL    time.Duration
	InterviewStateTTL time.Duration
}
