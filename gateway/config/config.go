package config

import "os"

// Config 网关服务配置，全部从环境变量读取。
type Config struct {
	HTTPAddr    string // 监听地址，默认 :8081
	EtcdAddr    string // etcd 地址，默认 localhost:2379
	InterviewServiceName string // 面试服务在 etcd 中的服务名
}

func Load() *Config {
	return &Config{
		HTTPAddr:             getEnv("GATEWAY_HTTP_ADDR", ":8081"),
		EtcdAddr:             getEnv("GATEWAY_ETCD_ADDR", "localhost:2379"),
		InterviewServiceName: getEnv("GATEWAY_INTERVIEW_SERVICE", "ai_interview"),
	}
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
