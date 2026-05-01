package handler

import "net/http"

type questionnaireHandler struct{}

// Get GET /v1/questionnaire?interview_id={}
// 获取面试结束后的问卷列表（每轮问题 + 用户回答）。
func (h *questionnaireHandler) Get(w http.ResponseWriter, r *http.Request) {
	panic("not implemented")
}

// Submit POST /v1/questionnaire/submit
// 提交用户对每轮对话的质量评价（good / bad）及文字反馈。
func (h *questionnaireHandler) Submit(w http.ResponseWriter, r *http.Request) {
	panic("not implemented")
}
