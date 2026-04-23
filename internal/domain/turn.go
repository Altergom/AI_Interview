package domain

import "time"

// InterviewTurn 对应 PostgreSQL interview_turns 与 Record Service。
type InterviewTurn struct {
	ID          string    `json:"id"`
	InterviewID string    `json:"interview_id"`
	TurnID      string    `json:"turn_id"`
	Stage       string    `json:"stage"`
	Question    string    `json:"question,omitempty"`
	UserAnswer  string    `json:"user_answer,omitempty"`
	ASRRaw      string    `json:"asr_raw,omitempty"`
	CreatedAt   time.Time `json:"created_at"`
}
