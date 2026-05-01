package handler

import "net/http"

type authHandler struct{}

// Register POST /v1/auth/register
func (h *authHandler) Register(w http.ResponseWriter, r *http.Request) {
	panic("not implemented")
}

// Login POST /v1/auth/login
func (h *authHandler) Login(w http.ResponseWriter, r *http.Request) {
	panic("not implemented")
}

// Guest POST /v1/auth/guest
// 生成临时游客账号，JWT 有效期 24h。
func (h *authHandler) Guest(w http.ResponseWriter, r *http.Request) {
	panic("not implemented")
}
