package handler

import (
	"encoding/json"
	"net/http"
)

// Resp 统一响应格式
type Resp struct {
	Code int `json:"code"`
	Data any `json:"data,omitempty"`
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

func ok(w http.ResponseWriter, data any) {
	writeJSON(w, http.StatusOK, Resp{Code: 0, Data: data})
}
