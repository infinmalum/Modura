// Package httpserver composes the public HTTP transport.
package httpserver

import (
	"context"
	"log/slog"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/modura-dev/modura/backend/internal/api/generated"
	"github.com/modura-dev/modura/backend/internal/api/handler"
	"github.com/modura-dev/modura/backend/internal/modules/identity"
	"github.com/modura-dev/modura/backend/internal/platform/config"
)

// Dependencies contains runtime application and health dependencies.
type Dependencies struct {
	Identity       handler.Identity
	Authorizer     handler.Authorizer
	Organization   handler.Organization
	PlatformAdmin  handler.PlatformAdmin
	PlatformTenant handler.PlatformTenant
	Provisioning   handler.Provisioning
	Ready          func(context.Context) error
}

// New returns a configured HTTP server without starting it.
func New(cfg config.HTTP, logger *slog.Logger, dependencies ...Dependencies) *http.Server {
	gin.SetMode(gin.ReleaseMode)
	router := gin.New()
	router.Use(gin.Recovery(), requestCorrelation(), requestLogger(logger))
	var deps Dependencies
	if len(dependencies) > 0 {
		deps = dependencies[0]
	}
	contractHandler := handler.New(handler.Dependencies{Identity: deps.Identity, Authorizer: deps.Authorizer, Organization: deps.Organization, PlatformAdmin: deps.PlatformAdmin, PlatformTenant: deps.PlatformTenant, Provisioning: deps.Provisioning, Ready: deps.Ready}, cfg.CookieSecure, func() (string, error) { return identity.NewOpaqueToken(32) })
	generated.RegisterHandlersWithOptions(router, contractHandler, generated.GinServerOptions{BaseURL: "/api"})
	return &http.Server{Addr: cfg.Address, Handler: router, ReadTimeout: cfg.ReadTimeout, WriteTimeout: cfg.WriteTimeout, IdleTimeout: cfg.IdleTimeout, MaxHeaderBytes: cfg.MaxHeaderBytes}
}

func requestCorrelation() gin.HandlerFunc {
	return func(c *gin.Context) {
		requestID := strings.TrimSpace(c.GetHeader("X-Request-ID"))
		if requestID == "" || len(requestID) > 128 {
			requestID = uuid.NewString()
			c.Request.Header.Set("X-Request-ID", requestID)
		}
		c.Header("X-Request-ID", requestID)
		c.Next()
	}
}

func requestLogger(logger *slog.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()
		logger.InfoContext(c.Request.Context(), "http request", "request_id", c.GetHeader("X-Request-ID"), "method", c.Request.Method, "path", c.Request.URL.Path, "status", c.Writer.Status())
	}
}
