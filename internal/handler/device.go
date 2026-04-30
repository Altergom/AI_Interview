package handler

import "net/http"

type deviceHandler struct{}

// Check POST /v1/device/check
// 前端完成权限获取后调用，服务端记录设备信息。
func (h *deviceHandler) Check(w http.ResponseWriter, r *http.Request) {
	panic("not implemented")
}
