package redis

import "fmt"

// Key 前缀与 技术文档「Redis」一节一致。

const (
	PrefixInterview = "interview"
	PrefixResume    = "resume"
)

// InterviewStateKey interview:{id}:state
func InterviewStateKey(interviewID string) string {
	return fmt.Sprintf("%s:%s:state", PrefixInterview, interviewID)
}

// InterviewConfigKey interview:{id}:config
// 存储面试配置（岗位方向等），由 SetConfig 写入，question_selector 读取。
func InterviewConfigKey(interviewID string) string {
	return fmt.Sprintf("%s:%s:config", PrefixInterview, interviewID)
}

// InterviewHistoryKey interview:{id}:history
func InterviewHistoryKey(interviewID string) string {
	return fmt.Sprintf("%s:%s:history", PrefixInterview, interviewID)
}

// InterviewAskedQuestionsKey interview:{id}:asked_questions
// 存储已出过的题目 hash（SHA-256 前 16 字节 hex），用于跨轮去重。
func InterviewAskedQuestionsKey(interviewID string) string {
	return fmt.Sprintf("%s:%s:asked_questions", PrefixInterview, interviewID)
}

// ResumeKey resume:{user_id}
func ResumeKey(userID string) string {
	return fmt.Sprintf("%s:%s", PrefixResume, userID)
}
