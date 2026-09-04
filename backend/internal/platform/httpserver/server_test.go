package httpserver

import (
	"context"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/modura-dev/modura/backend/internal/api/generated"
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

func TestGeneratedParameterErrorsAreNonLeakingProblems(t *testing.T) {
	cfg := config.HTTP{Address: ":0", ReadTimeout: time.Second, WriteTimeout: time.Second, IdleTimeout: time.Second, ShutdownTimeout: time.Second, MaxHeaderBytes: 1024}
	server := New(cfg, slog.New(slog.NewTextHandler(io.Discard, nil)))
	request := httptest.NewRequestWithContext(context.Background(), http.MethodPost, "/api/platform/tenants/not-a-uuid/suspend", nil)
	response := httptest.NewRecorder()
	server.Handler.ServeHTTP(response, request)
	if response.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d", response.Code, http.StatusBadRequest)
	}
	if response.Header().Get("Content-Type") != "application/problem+json" {
		t.Fatalf("content type = %q", response.Header().Get("Content-Type"))
	}
	var problem generated.Problem
	if err := json.NewDecoder(response.Body).Decode(&problem); err != nil {
		t.Fatal(err)
	}
	if problem.Title != "invalid request" || problem.Status != http.StatusBadRequest {
		t.Fatalf("problem = %+v", problem)
	}
}
