package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/joho/godotenv"

	"ai_interview/internal/mq"
)

// Default 搜索路径按顺序尝试加载，先加载的文件先生效；后加载的 .env 可覆盖前文件中的键
// （与 godotenv 行为：后 Load 会覆盖同键）。通常本地使用 `.env` + 可选 `.env.local` 覆盖。
var defaultEnvFiles = []string{".env", ".env.local"}
var Cfg *Config

// Load 从 .env 文件与进程环境构建 Config。若未设置某变量则使用合理默认值；解析失败时返回错误。
// 可传入自定义 env 文件路径，传 nil 或空则使用 defaultEnvFiles。
func Load(envFiles ...string) error {
	var err error

	files := envFiles
	if len(files) == 0 {
		files = defaultEnvFiles
	}
	for _, f := range files {
		_ = godotenv.Load(f) // 文件不存在不报错
	}

	Cfg, err = parseFromEnv()
	if err != nil {
		return fmt.Errorf("[load]config", "err", err)
	}

	return nil
}

func get(key, def string) string {
	v := strings.TrimSpace(os.Getenv(key))
	if v == "" {
		return def
	}
	return v
}

func getInt(key string, def int) (int, error) {
	s := strings.TrimSpace(os.Getenv(key))
	if s == "" {
		return def, nil
	}
	i, err := strconv.Atoi(s)
	if err != nil {
		return 0, fmt.Errorf("config %q: %w", key, err)
	}
	return i, nil
}

func getBool(key string, def bool) (bool, error) {
	s := strings.TrimSpace(strings.ToLower(os.Getenv(key)))
	if s == "" {
		return def, nil
	}
	b, err := strconv.ParseBool(s)
	if err != nil {
		return false, fmt.Errorf("config %q: %w", key, err)
	}
	return b, nil
}

// splitComma 按逗号分割并 trim 空白，用于多地址配置（如 ES_ADDRS）。
func splitComma(s string) []string {
	parts := strings.Split(s, ",")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		if v := strings.TrimSpace(p); v != "" {
			out = append(out, v)
		}
	}
	return out
}

func getDurationOrDefault(key string, def time.Duration) (time.Duration, error) {
	s := strings.TrimSpace(os.Getenv(key))
	if s == "" {
		return def, nil
	}
	d, err := time.ParseDuration(s)
	if err != nil {
		return 0, fmt.Errorf("config %q: %w", key, err)
	}
	return d, nil
}

func parseFromEnv() (*Config, error) {
	resumeTTL, err := getDurationOrDefault("RESUME_REDIS_TTL", ResumeRedisTTL)
	if err != nil {
		return nil, err
	}
	interviewTTL, err := getDurationOrDefault("INTERVIEW_STATE_TTL", InterviewStateTTL)
	if err != nil {
		return nil, err
	}

	redisDB, err := getInt("REDIS_DB", 0)
	if err != nil {
		return nil, err
	}

	accessMin, err := getInt("JWT_ACCESS_EXP_MIN", 60)
	if err != nil {
		return nil, err
	}

	s3ssl, err := getBool("S3_USE_SSL", true)
	if err != nil {
		return nil, err
	}

	logFormat := strings.ToLower(strings.TrimSpace(get("LOG_FORMAT", "text")))
	if logFormat != "json" && logFormat != "text" {
		logFormat = "text"
	}

	return &Config{
		Env:         get("APP_ENV", "development"),
		LogLevel:    get("LOG_LEVEL", "info"),
		LogFormat:   logFormat,
		HTTPAddr:    get("HTTP_ADDR", ":8080"),
		PostgresDSN: get("POSTGRES_DSN", ""),

		RedisAddr:     get("REDIS_ADDR", "127.0.0.1:6379"),
		RedisPassword: get("REDIS_PASSWORD", ""),
		RedisDB:       redisDB,

		MQBrokerURL:              get("MQ_BROKER_URL", ""),
		MQTopicInterviewFinished: get("MQ_TOPIC_INTERVIEW_FINISHED", mq.TopicInterviewFinished),

		S3Endpoint:  get("S3_ENDPOINT", ""),
		S3AccessKey: get("S3_ACCESS_KEY", ""),
		S3SecretKey: get("S3_SECRET_KEY", ""),
		S3Bucket:    get("S3_BUCKET", ""),
		S3Region:    get("S3_REGION", ""),
		S3UseSSL:    s3ssl,

		JWTSecret:       get("JWT_SECRET", ""),
		JWTIssuer:       get("JWT_ISSUER", "ai_interview"),
		JWTAccessExpMin: accessMin,

		OpenAIAPIKey:    get("OPENAI_API_KEY", ""),
		OpenAIBaseURL:   get("OPENAI_BASE_URL", ""),
		DoubaoAPIKey:    get("DOUBAO_API_KEY", ""),
		DoubaoBaseURL:   get("DOUBAO_BASE_URL", ""),
		QwenAPIKey:      get("QWEN_API_KEY", ""),
		QwenBaseURL:     get("QWEN_BASE_URL", ""),
		ClaudeAPIKey:    get("CLAUDE_API_KEY", ""),
		ClaudeBaseURL:   get("CLAUDE_BASE_URL", ""),
		DeepSeekAPIKey:  get("DEEPSEEK_API_KEY", ""),
		DeepSeekBaseURL: get("DEEPSEEK_BASE_URL", ""),
		GenimiAPIKey:    get("GEMINI_API_KEY", ""),
		GenimiBaseURL:   get("GEMINI_BASE_URL", ""),

		Supervisor: get("SUPERVISOR", "qwen-plus"),
		Selector:   get("SELECTOR", "qwen-plus"),
		Manager:    get("MANAGER", "qwen-plus"),
		Analyzer:   get("ANALYZER", "qwen-plus"),
		Evaluator:  get("EVALUATOR", "qwen-plus"),

		ASRModel: get("ASR_MODEL", "qwen3-asr-flash-realtime"),
		TTSModel: get("TTS_MODEL", "qwen-tts-flash-realtime"),
		TTSVoice: get("TTS_VOICE", "zhifeng_emo"),

		MilvusAddr:       get("MILVUS_ADDR", "127.0.0.1:19530"),
		MilvusCollection: get("MILVUS_COLLECTION", "bank_questions_vec"),

		ESAddrs:    splitComma(get("ES_ADDRS", "http://127.0.0.1:9200")),
		ESUsername: get("ES_USERNAME", ""),
		ESPassword: get("ES_PASSWORD", ""),
		ESIndex:    get("ES_INDEX", "bank_questions"),

		OllamaBaseURL:    get("OLLAMA_BASE_URL", "http://127.0.0.1:11434"),
		OllamaEmbedModel: get("OLLAMA_EMBED_MODEL", "nomic-embed-text"),
		OllamaChatModel:  get("OLLAMA_CHAT_MODEL", "qwen3:8b"),

		ResumeRedisTTL:    resumeTTL,
		InterviewStateTTL: interviewTTL,
	}, nil
}
