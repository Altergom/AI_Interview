package health

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"ai_interview/internal/config"
)

func TestLive(t *testing.T) {
	h := New(&config.App{})
	mux := http.NewServeMux()
	h.Register(mux)
	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("code %d", rec.Code)
	}
}
