// Package httpserver composes the public HTTP transport.
package httpserver

import (
	"context"
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/modura-dev/modura/backend/internal/api/generated"
	"github.com/modura-dev/modura/backend/internal/api/handler"
	"github.com/modura-dev/modura/backend/internal/modules/identity"
	"github.com/modura-dev/modura/backend/internal/platform/config"
)

// Dependencies contains runtime application and health dependencies.
type Dependencies struct {
	Identity handler.Identity
	Ready    func(context.Context) error
}

// New returns a configured HTTP server without starting it.
func New(cfg config.HTTP, logger *slog.Logger, dependencies ...Dependencies) *http.Server {
	gin.SetMode(gin.ReleaseMode)
	router := gin.New()
	router.Use(gin.Recovery(), requestLogger(logger))
	var deps Dependencies
	if len(dependencies) > 0 {
		deps = dependencies[0]
	}
	contractHandler := handler.New(deps.Identity, cfg.CookieSecure, func() (string, error) { return identity.NewOpaqueToken(32) }, deps.Ready)
	generated.RegisterHandlersWithOptions(router, contractHandler, generated.GinServerOptions{BaseURL: "/api"})
	return &http.Server{Addr: cfg.Address, Handler: router, ReadTimeout: cfg.ReadTimeout, WriteTimeout: cfg.WriteTimeout, IdleTimeout: cfg.IdleTimeout, MaxHeaderBytes: cfg.MaxHeaderBytes}
}

func requestLogger(logger *slog.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()
		logger.InfoContext(c.Request.Context(), "http request", "method", c.Request.Method, "path", c.Request.URL.Path, "status", c.Writer.Status())
	}
}
