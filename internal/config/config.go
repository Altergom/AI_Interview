package config

import "time"

// Config 从环境变量与 .env 文件加载的集中配置
type Config struct {
	Env       string
	LogLevel  string
	LogFormat string // json | text；json 便于接入 Loki / ELK / 云厂商日志采集

	// HTTP（网关 / 健康检查等）
	HTTPAddr string

	// PostgreSQL
	PostgresDSN       string
	PGMaxOpenConns    int           // 最大打开连接数，默认 25
	PGMaxIdleConns    int           // 最大空闲连接数，默认 5
	PGConnMaxLifetime time.Duration // 连接最大存活时间，默认 30m
	PGConnMaxIdleTime time.Duration // 连接最大空闲时间，默认 5m

	// Redis
	RedisAddr         string
	RedisPassword     string
	RedisDB           int
	RedisPoolSize     int           // 连接池大小，默认 10
	RedisMinIdleConns int           // 最小空闲连接数，默认 2
	RedisDialTimeout  time.Duration // 连接超时，默认 5s
	RedisReadTimeout  time.Duration // 读超时，默认 3s
	RedisWriteTimeout time.Duration // 写超时，默认 3s

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

	// 模型提供商基本配置
	OpenAIAPIKey    string
	OpenAIBaseURL   string
	DoubaoAPIKey    string
	DoubaoBaseURL   string
	QwenAPIKey      string
	QwenBaseURL     string
	ClaudeAPIKey    string
	ClaudeBaseURL   string
	DeepSeekAPIKey  string
	DeepSeekBaseURL string
	GenimiAPIKey    string
	GenimiBaseURL   string

	// 模型
	Supervisor string
	Selector   string
	Manager    string
	Analyzer   string
	Evaluator  string

	// ASR/TTS 模型配置
	ASRModel string // ASR 模型名称
	TTSModel string // TTS 模型名称
	TTSVoice string // TTS 音色

	// Milvus 向量数据库
	MilvusAddr       string // host:port，默认 127.0.0.1:19530
	MilvusCollection string // 题库向量集合名，默认 bank_questions_vec
	MilvusAPIKey     string //
	MilvusEnableTLS  bool   // Zilliz Cloud 必须 true
	
	// Elasticsearch
	ESAddrs    []string // 节点地址列表，逗号分隔
	ESUsername string
	ESPassword string
	ESIndex    string // 题库 ES 索引名，默认 bank_questions

	// Ollama（本地模型，可选）
	OllamaBaseURL    string // 默认 http://127.0.0.1:11434
	OllamaEmbedModel string // embedding 模型，默认 nomic-embed-text
	OllamaChatModel  string // chat 模型，默认 qwen3:8b

	// TTL：可由 RESUME_REDIS_TTL、INTERVIEW_STATE_TTL 等 duration 字符串覆盖
	ResumeRedisTTL    time.Duration
	InterviewStateTTL time.Duration

	// 面试流程驱动模式
	// WorkflowEnabled=true 使用显式 workflow（StageRouter + 状态机裁决）；
	// WorkflowEnabled=false 使用 Agent Supervisor 驱动模式。
	WorkflowEnabled bool

	// Skill 出题模块
	// SkillsDir SKILL.md 所在父目录的绝对路径，默认 "internal/einocore/skills"（相对项目根）
	SkillsDir string
}
