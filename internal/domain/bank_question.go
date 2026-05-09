package domain

import "time"

// VecStatus 向量化异步状态。
type VecStatus string

const (
	VecStatusPending VecStatus = "pending"
	VecStatusDone    VecStatus = "done"
	VecStatusFailed  VecStatus = "failed"
)

// Difficulty 题目难度。
type Difficulty string

const (
	DifficultyEasy   Difficulty = "easy"
	DifficultyMedium Difficulty = "medium"
	DifficultyHard   Difficulty = "hard"
)

// BankQuestion 八股题库单条（RAG 索引用，与 TODO 一致）。
type BankQuestion struct {
	QuestionID     string   `json:"question_id"`
	Question       string   `json:"question"`
	Tags           []string `json:"tags"`
	StandardAnswer string   `json:"standard_answer"`
	FollowUpHints  []string `json:"follow_up_hints"`
}

// BankQuestionRecord PgSQL 完整记录，包含元数据字段。
type BankQuestionRecord struct {
	ID                   string     `json:"id"`
	Question             string     `json:"question"`
	StandardAnswer       string     `json:"standard_answer"`
	Tags                 []string   `json:"tags"`
	RelatedConcepts      []string   `json:"related_concepts"`
	FollowupQuestionIDs  []string   `json:"followup_question_ids"`
	Difficulty           Difficulty `json:"difficulty"`
	VecStatus            VecStatus  `json:"vec_status"`
	CreatedAt            time.Time  `json:"created_at"`
	UpdatedAt            time.Time  `json:"updated_at"`
}

// RetrievedQuestion 召回结果，附带融合分数。
type RetrievedQuestion struct {
	BankQuestionRecord
	Score float64 `json:"score"`
}
