package domain

import "time"

// InterviewState 存 Redis key: interview:{interview_id}:state
type InterviewState struct {
	InterviewID              string         `json:"interview_id"`
	Stage                    InterviewStage `json:"stage"`
	Direction                string         `json:"direction"`  // 面试方向：go-backend / java-backend / frontend / algorithm / ai-agent
	Position                 string         `json:"position"`   // 岗位名称（可选，用于 prompt 上下文）
	QuestionsAsked           int            `json:"questions_asked"`
	CurrentQuestionFollowups int            `json:"current_question_followups"`
	ReportStatus             string         `json:"report_status"` // pending | processing | completed | failed
	StartedAt                time.Time      `json:"started_at"`
}

// InterviewSession 面试会话，存 Redis key: interview:session:{interview_id}
type InterviewSession struct {
	InterviewID string         `json:"interview_id"`
	UserID      string         `json:"user_id"`
	Stage       InterviewStage `json:"stage"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	History     []SessionMessage `json:"history"`
	Stats       SessionStats   `json:"stats"`
	Context     map[string]any `json:"context"`
}

// SessionMessage 单条对话消息。
type SessionMessage struct {
	Role      string    `json:"role"`      // user | assistant
	Content   string    `json:"content"`
	Timestamp time.Time `json:"timestamp"`
}

// SessionStats 会话统计信息。
type SessionStats struct {
	QuestionCount  int `json:"question_count"`
	AlgorithmCount int `json:"algorithm_count"`
	TotalRounds    int `json:"total_rounds"`
}
