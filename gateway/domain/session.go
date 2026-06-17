package domain

import "time"

type SessionStatus string

const (
	StatusNew          SessionStatus = "new"
	StatusVerifying    SessionStatus = "verifying"
	StatusReady        SessionStatus = "ready"
	StatusInterviewing SessionStatus = "interviewing"
	StatusWaitingUser  SessionStatus = "waiting_user"
	StatusPaused       SessionStatus = "paused"
	StatusScoring      SessionStatus = "scoring"
	StatusFinished     SessionStatus = "finished"
	StatusHandoff      SessionStatus = "handoff"
	StatusExpired      SessionStatus = "expired"
)

// GatewaySession 网关层面试会话，记录渠道身份与状态机节点。
// 不存面试内容，面试内容在面试服务侧通过 InterviewID 关联查询。
type GatewaySession struct {
	SessionID    string
	CandidateID  string
	Channel      string // wechat / feishu / qqbot
	PeerID       string // 用户在平台的唯一 ID
	Status       SessionStatus
	InterviewID  string // 关联面试服务的 interview_id
	CreatedAt    time.Time
	UpdatedAt    time.Time
	LastActiveAt time.Time
}
