package mq

import "time"

// TopicInterviewFinished 默认 topic 名称。
// 业务代码应从 config.App.MQTopicInterviewFinished 读取（支持环境变量覆盖），不直接引用此常量。
const TopicInterviewFinished = "interview_finished"

// EventNameInterviewFinished 消息体中的 event 字段值。
const EventNameInterviewFinished = "interview_finished"

// InterviewFinished 消息体（技术文档 / TODO）。
type InterviewFinished struct {
	Event       string    `json:"event"`
	InterviewID string    `json:"interview_id"`
	UserID      string    `json:"user_id"`
	FinishedAt  time.Time `json:"finished_at"`
}

func NewInterviewFinishedEvent(interviewID, userID string, t time.Time) InterviewFinished {
	return InterviewFinished{
		Event:       EventNameInterviewFinished,
		InterviewID: interviewID,
		UserID:      userID,
		FinishedAt:  t,
	}
}
