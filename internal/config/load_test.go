package config

import (
	"testing"
	"time"
)

func TestLoadDefaults(t *testing.T) {
	// 不读 .env 文件，仅验证环境变量缺省与 parse 逻辑
	for _, k := range []string{
		"APP_ENV", "RESUME_REDIS_TTL", "INTERVIEW_STATE_TTL", "MQ_TOPIC_INTERVIEW_FINISHED", "LOG_FORMAT",
	} {
		t.Setenv(k, "")
	}
	cfg, err := parseFromEnv()
	if err != nil {
		t.Fatal(err)
	}
	if cfg.Env != "development" {
		t.Fatalf("Env: got %q", cfg.Env)
	}
	if cfg.ResumeRedisTTL != 7*24*time.Hour {
		t.Fatalf("ResumeRedisTTL: got %v", cfg.ResumeRedisTTL)
	}
	if cfg.MQTopicInterviewFinished != "interview_finished" {
		t.Fatalf("topic: got %q", cfg.MQTopicInterviewFinished)
	}
	if cfg.LogFormat != "text" {
		t.Fatalf("LogFormat: got %q", cfg.LogFormat)
	}
}

func TestLoadDurationOverride(t *testing.T) {
	t.Setenv("RESUME_REDIS_TTL", "72h")
	t.Setenv("INTERVIEW_STATE_TTL", "24h")
	cfg, err := parseFromEnv()
	if err != nil {
		t.Fatal(err)
	}
	if cfg.ResumeRedisTTL != 72*time.Hour {
		t.Fatalf("ResumeRedisTTL: got %v", cfg.ResumeRedisTTL)
	}
	if cfg.InterviewStateTTL != 24*time.Hour {
		t.Fatalf("InterviewStateTTL: got %v", cfg.InterviewStateTTL)
	}
}
