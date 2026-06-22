package handler

import (
	"context"
	"net/http"

	"github.com/cloudwego/hertz/pkg/app"

	"ai_interview/internal/domain"
	authmw "ai_interview/internal/middleware/auth"
	"ai_interview/internal/service"
	biz "ai_interview/internal/utils/respx"
)

type questionnaireHandler struct {
	svc service.QuestionnaireService
}

// Get GET /v1/questionnaire?interview_id=xxx
// 返回该面试的所有 turn，供前端渲染问卷条目。
func (h *questionnaireHandler) Get(ctx context.Context, c *app.RequestContext) {
	userID := authmw.GetUserID(c)
	interviewID := string(c.Query("interview_id"))
	if interviewID == "" {
		Fail(ctx, c, http.StatusBadRequest, biz.CodeBadRequest, "interview_id is required")
		return
	}

	turns, err := h.svc.Get(ctx, userID, interviewID)
	if err != nil {
		HandleErr(ctx, c, err)
		return
	}
	OK(ctx, c, turns)
}

// submitAnswerReq 单条标注的请求体 DTO，与 domain 解耦。
type submitAnswerReq struct {
	TurnID   string `json:"turn_id"`
	Quality  string `json:"quality"`
	Feedback string `json:"feedback,omitempty"`
}

// submitQuestionnaireReq POST /v1/questionnaire/submit 请求体。
type submitQuestionnaireReq struct {
	InterviewID string            `json:"interview_id"`
	Answers     []submitAnswerReq `json:"answers"`
}

// Submit POST /v1/questionnaire/submit
// 批量提交问卷标注，good/bad 都采集（DPO 负样本用）。
// 允许部分提交：未标注的 turn 不出现在 answers 中即可。
func (h *questionnaireHandler) Submit(ctx context.Context, c *app.RequestContext) {
	userID := authmw.GetUserID(c)

	var req submitQuestionnaireReq
	if err := c.BindJSON(&req); err != nil {
		Fail(ctx, c, http.StatusBadRequest, biz.CodeBadRequest, "invalid request body")
		return
	}

	answers := make([]service.QuestionnaireAnswer, len(req.Answers))
	for i, a := range req.Answers {
		answers[i] = service.QuestionnaireAnswer{
			TurnID:   a.TurnID,
			Quality:  domain.QuestionnaireQuality(a.Quality),
			Feedback: a.Feedback,
		}
	}

	if err := h.svc.Submit(ctx, service.QuestionnaireSubmitRequest{
		UserID:      userID,
		InterviewID: req.InterviewID,
		Answers:     answers,
	}); err != nil {
		HandleErr(ctx, c, err)
		return
	}

	OK(ctx, c, map[string]any{"submitted": len(answers)})
}
