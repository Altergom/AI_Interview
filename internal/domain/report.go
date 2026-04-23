package domain

import "time"

// Report 对应 PostgreSQL reports 与 Report Worker 输出。
type Report struct {
	ID             string    `json:"id"`
	InterviewID    string    `json:"interview_id"`
	KnowledgeDepth int       `json:"knowledge_depth"`
	Expression     int       `json:"expression"`
	ProblemSolving int       `json:"problem_solving"`
	CodeQuality    int       `json:"code_quality"`
	StressResponse int       `json:"stress_response"`
	Summary        string    `json:"summary"`
	WeakPoints     []string  `json:"weak_points"`
	StrongPoints   []string  `json:"strong_points"`
	CreatedAt      time.Time `json:"created_at"`
}

// ReportDimensions 技术文档中的 JSON 视图。
type ReportDimensions struct {
	KnowledgeDepth int `json:"knowledge_depth"`
	Expression     int `json:"expression"`
	ProblemSolving int `json:"problem_solving"`
	CodeQuality    int `json:"code_quality"`
	StressResponse int `json:"stress_response"`
}
