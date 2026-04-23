package domain

// BankQuestion 八股题库单条（RAG 索引用，与 TODO 一致）。
type BankQuestion struct {
	QuestionID     string   `json:"question_id"`
	Question       string   `json:"question"`
	Tags           []string `json:"tags"`
	StandardAnswer string   `json:"standard_answer"`
	FollowUpHints  []string `json:"follow_up_hints"`
}
