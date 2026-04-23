package domain

// CodeJudgeResult Code Judge Agent 输出（Eino 节点间传递用）。
type CodeJudgeResult struct {
	Correctness     bool     `json:"correctness"`
	TimeComplexity  string   `json:"time_complexity"`
	SpaceComplexity string   `json:"space_complexity"`
	Issues          []string `json:"issues"`
}
