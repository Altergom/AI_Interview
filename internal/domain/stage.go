package domain

// InterviewStage 对应 Router 中的阶段（与 TODO / 技术文档一致）。
type InterviewStage string

const (
	StageIntro       InterviewStage = "intro"
	StageQuestioning InterviewStage = "questioning"
	StageAlgorithm   InterviewStage = "algorithm"
	StageClosing     InterviewStage = "closing"
	StageEnd         InterviewStage = "end"
)

func (s InterviewStage) String() string { return string(s) }
