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

// InterviewHistoryKey interview:{id}:history
func InterviewHistoryKey(interviewID string) string {
	return fmt.Sprintf("%s:%s:history", PrefixInterview, interviewID)
}

// ResumeKey resume:{user_id}
func ResumeKey(userID string) string {
	return fmt.Sprintf("%s:%s", PrefixResume, userID)
}

// InterviewConfigKey interview:config:{config_id}
func InterviewConfigKey(configID string) string {
	return fmt.Sprintf("%s:config:%s", PrefixInterview, configID)
}

// UserInterviewConfigKey interview:user:{user_id}:config
func UserInterviewConfigKey(userID string) string {
	return fmt.Sprintf("%s:user:%s:config", PrefixInterview, userID)
}
