package mq

import "time"

// TopicInterviewFinished 默认 topic 名称。
// 业务代码应从 config.App.MQTopicInterviewFinished 读取（支持环境变量覆盖），不直接引用此常量。
const TopicInterviewFinished = "interview_finished"

// TopicVectorizeTask 题目向量化任务队列。
// 题目插入 PG 后发布此消息，Worker 异步写 Milvus+ES 并更新 vec_status。
const TopicVectorizeTask = "vectorize_task"

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

// VectorizeTask 向量化任务消息体。
// 题目写入 PG 后 Producer 投递，Worker 拉取并完成 Milvus+ES 写入。
type VectorizeTask struct {
	QuestionID string `json:"question_id"`
}
