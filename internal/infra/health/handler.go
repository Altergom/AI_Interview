package health

import (
	"context"
	"encoding/json"
	"net/http"
	"os"
	"time"

	"ai_interview/internal/config"
)

// Handler 提供存活与就绪探针，供 Kubernetes / Docker Compose / 负载均衡使用。
type Handler struct {
	cfg *config.App
}

func New(cfg *config.App) *Handler {
	return &Handler{cfg: cfg}
}

// Register 使用 Go 1.22+ 方法模式路由；探针路径与常见约定一致。
func (h *Handler) Register(mux *http.ServeMux) {
	mux.HandleFunc("GET /health", h.live)
	mux.HandleFunc("GET /healthz", h.live)
	mux.HandleFunc("GET /ready", h.ready)
	mux.HandleFunc("GET /readyz", h.ready)
}

func (h *Handler) live(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
}

func (h *Handler) ready(w http.ResponseWriter, r *http.Request) {
	if os.Getenv("READINESS_SKIP_ALL") == "1" {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"status":  "ready",
			"skipped": true,
		})
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	checks, ok := runReadiness(ctx, h.cfg)
	w.Header().Set("Content-Type", "application/json")
	if !ok {
		w.WriteHeader(http.StatusServiceUnavailable)
	}
	_ = json.NewEncoder(w).Encode(map[string]any{
		"status": map[bool]string{true: "ready", false: "not_ready"}[ok],
		"checks": checks,
	})
}
