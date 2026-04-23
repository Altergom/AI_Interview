package config

import "time"

// 与技术文档中 Redis TTL 建议一致，后续可从环境变量覆盖。

const (
	ResumeRedisTTL    = 7 * 24 * time.Hour
	InterviewStateTTL = 48 * time.Hour // 面试进行中超时兜底，业务结束时应主动删除
)
