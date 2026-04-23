package domain

import "time"

// InterviewState 存 Redis key: interview:{interview_id}:state
type InterviewState struct {
	InterviewID              string         `json:"interview_id"`
	Stage                    InterviewStage `json:"stage"`
	QuestionsAsked           int            `json:"questions_asked"`
	CurrentQuestionFollowups int            `json:"current_question_followups"`
	StartedAt                time.Time      `json:"started_at"`
}
