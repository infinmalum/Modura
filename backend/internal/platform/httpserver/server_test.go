package httpserver

import (
	"context"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/modura-dev/modura/backend/internal/platform/config"
)

func TestHealthEndpoints(t *testing.T) {
	cfg := config.HTTP{Address: ":0", ReadTimeout: time.Second, WriteTimeout: time.Second, IdleTimeout: time.Second, ShutdownTimeout: time.Second, MaxHeaderBytes: 1024}
	server := New(cfg, slog.New(slog.NewTextHandler(io.Discard, nil)))
	for _, path := range []string{"/api/livez", "/api/readyz"} {
		recorder := httptest.NewRecorder()
		request := httptest.NewRequestWithContext(context.Background(), http.MethodGet, path, nil)
		server.Handler.ServeHTTP(recorder, request)
		if recorder.Code != http.StatusOK {
			t.Fatalf("GET %s status = %d, want %d", path, recorder.Code, http.StatusOK)
		}
	}
}
