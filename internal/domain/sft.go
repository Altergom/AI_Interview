package domain

// SFTMessage JSONL 训练数据中的单条消息（与 Eino/OpenAI schema 对齐的 role/content）。
type SFTMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// SFTJSONLRecord 问卷 job 导出到 S3 的 JSONL 行结构（见 TODO）。
type SFTJSONLRecord struct {
	Messages     []SFTMessage `json:"messages"`
	Quality      string       `json:"quality"`
	UserFeedback string       `json:"user_feedback"`
}
