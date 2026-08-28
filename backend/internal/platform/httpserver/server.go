// Package httpserver composes the public HTTP transport.
package httpserver

import (
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/modura-dev/modura/backend/internal/api/generated"
	"github.com/modura-dev/modura/backend/internal/platform/config"
)

type healthHandler struct{}

// New returns a configured HTTP server without starting it.
func New(cfg config.HTTP, logger *slog.Logger) *http.Server {
	gin.SetMode(gin.ReleaseMode)
	router := gin.New()
	router.Use(gin.Recovery(), requestLogger(logger))
	generated.RegisterHandlersWithOptions(router, healthHandler{}, generated.GinServerOptions{BaseURL: "/api"})
	return &http.Server{Addr: cfg.Address, Handler: router, ReadTimeout: cfg.ReadTimeout, WriteTimeout: cfg.WriteTimeout, IdleTimeout: cfg.IdleTimeout, MaxHeaderBytes: cfg.MaxHeaderBytes}
}

func (healthHandler) GetLiveness(c *gin.Context) {
	c.JSON(http.StatusOK, generated.HealthStatus{Status: generated.Ok})
}

func (healthHandler) GetReadiness(c *gin.Context) {
	c.JSON(http.StatusOK, generated.HealthStatus{Status: generated.Ok})
}

func requestLogger(logger *slog.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()
		logger.InfoContext(c.Request.Context(), "http request", "method", c.Request.Method, "path", c.Request.URL.Path, "status", c.Writer.Status())
	}
}
