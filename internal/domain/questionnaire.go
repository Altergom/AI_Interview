package domain

import "time"

// QuestionnaireQuality 问卷打标。
type QuestionnaireQuality string

const (
	QualityGood QuestionnaireQuality = "good"
	QualityBad  QuestionnaireQuality = "bad"
)

// QuestionnaireResult 对应 PostgreSQL questionnaire_results。
type QuestionnaireResult struct {
	ID          string               `json:"id"`
	InterviewID string               `json:"interview_id"`
	TurnID      string               `json:"turn_id"`
	Quality     QuestionnaireQuality `json:"quality"`
	Feedback    string               `json:"feedback,omitempty"`
	CreatedAt   time.Time            `json:"created_at"`
}
