package handler

import "net/http"

type reportHandler struct{}

// Status GET /v1/report/status?interview_id={}
// 查询报告生成状态：pending / generating / done / failed。
func (h *reportHandler) Status(w http.ResponseWriter, r *http.Request) {
	panic("not implemented")
}

// Get GET /v1/report?interview_id={}
// 获取已生成的报告，含各维度评分、总结、优劣势。
func (h *reportHandler) Get(w http.ResponseWriter, r *http.Request) {
	panic("not implemented")
}
