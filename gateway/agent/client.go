package agent

import "context"

// CreateInterviewReq CreateInterview 入参。
type CreateInterviewReq struct {
	CandidateID string
	Position    string
	Direction   string
}

// CreateInterviewResp CreateInterview 出参。
type CreateInterviewResp struct {
	InterviewID string
	Stage       string
	CreatedAt   string
}

// SubmitTurnReq SubmitTurn 入参。
type SubmitTurnReq struct {
	InterviewID string
	CandidateID string
	Text        string
}

// SubmitTurnResp SubmitTurn 出参。
type SubmitTurnResp struct {
	Reply      string
	Stage      string
	IsFinished bool
}

// FinishInterviewReq FinishInterview 入参。
type FinishInterviewReq struct {
	InterviewID string
}

// FinishInterviewResp FinishInterview 出参。
type FinishInterviewResp struct {
	FinishedAt      string
	DurationSeconds int64
}

// Client 封装对面试服务的 Kitex RPC 调用。
type Client struct {
	// TODO: 注入 Kitex 生成的 InterviewService client
}

func NewClient() *Client {
	return &Client{}
}

func (c *Client) CreateInterview(ctx context.Context, req *CreateInterviewReq) (*CreateInterviewResp, error) {
	// TODO: 调用 Kitex client.CreateInterview
	return &CreateInterviewResp{}, nil
}

func (c *Client) SubmitTurn(ctx context.Context, req *SubmitTurnReq) (*SubmitTurnResp, error) {
	// TODO: 调用 Kitex client.SubmitTurn
	return &SubmitTurnResp{}, nil
}

func (c *Client) FinishInterview(ctx context.Context, req *FinishInterviewReq) (*FinishInterviewResp, error) {
	// TODO: 调用 Kitex client.FinishInterview
	return &FinishInterviewResp{}, nil
}
