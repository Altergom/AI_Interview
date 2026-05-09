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
