// Package main composes and runs the Modura backend.
package main

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/modura-dev/modura/backend/internal/modules/audit"
	auditpostgres "github.com/modura-dev/modura/backend/internal/modules/audit/postgres"
	"github.com/modura-dev/modura/backend/internal/modules/authorization"
	authorizationpostgres "github.com/modura-dev/modura/backend/internal/modules/authorization/postgres"
	"github.com/modura-dev/modura/backend/internal/modules/identity"
	identitypostgres "github.com/modura-dev/modura/backend/internal/modules/identity/postgres"
	"github.com/modura-dev/modura/backend/internal/modules/organization"
	organizationpostgres "github.com/modura-dev/modura/backend/internal/modules/organization/postgres"
	"github.com/modura-dev/modura/backend/internal/modules/platformadmin"
	platformadminpostgres "github.com/modura-dev/modura/backend/internal/modules/platformadmin/postgres"
	"github.com/modura-dev/modura/backend/internal/modules/platformtenant"
	platformtenantpostgres "github.com/modura-dev/modura/backend/internal/modules/platformtenant/postgres"
	"github.com/modura-dev/modura/backend/internal/modules/provisioning"
	"github.com/modura-dev/modura/backend/internal/platform/config"
	"github.com/modura-dev/modura/backend/internal/platform/database"
	"github.com/modura-dev/modura/backend/internal/platform/httpserver"
	"github.com/modura-dev/modura/backend/internal/platform/identifier"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	cfg, err := config.FromEnv()
	if err != nil {
		logger.Error("load configuration", "error", err)
		os.Exit(1)
	}

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	startupCtx, cancelStartup := context.WithTimeout(ctx, 10*time.Second)
	defer cancelStartup()
	pool, err := pgxpool.New(startupCtx, cfg.Database.URL)
	if err != nil {
		logger.Error("configure database", "error", err)
		os.Exit(1)
	}
	defer pool.Close()
	if err := pool.Ping(startupCtx); err != nil {
		logger.Error("connect database", "error", err)
		os.Exit(1)
	}

	signer, err := identity.NewAccessTokenSigner(cfg.Auth.Issuer, cfg.Auth.Audience, cfg.Auth.SigningKeyID, cfg.Auth.SigningKey, cfg.Auth.AccessLifetime)
	if err != nil {
		logger.Error("configure access tokens", "error", err)
		os.Exit(1)
	}
	verifier := identity.NewAccessTokenVerifier(cfg.Auth.Issuer, cfg.Auth.Audience, map[string][]byte{cfg.Auth.SigningKeyID: cfg.Auth.SigningKey}, 5*time.Second)
	identityService, err := identity.NewService(identitypostgres.New(pool), signer, verifier, identity.DefaultPasswordParameters(), cfg.Auth.RefreshLifetime, time.Now, func(now time.Time) (string, error) {
		id, idErr := identifier.NewUUIDv7(now, nil)
		return string(id), idErr
	}, func() (string, error) { return identity.NewOpaqueToken(32) })
	if err != nil {
		logger.Error("configure identity service", "error", err)
		os.Exit(1)
	}
	authorizationService, err := authorization.NewService(authorizationpostgres.New(pool))
	if err != nil {
		logger.Error("configure authorization service", "error", err)
		os.Exit(1)
	}
	auditService, err := audit.NewService(auditpostgres.New(), func(now time.Time) (string, error) {
		id, idErr := identifier.NewUUIDv7(now, nil)
		return string(id), idErr
	})
	if err != nil {
		logger.Error("configure audit service", "error", err)
		os.Exit(1)
	}
	if err := authorizationService.EnableManagement(authorizationpostgres.New(pool), database.NewTransactor(pool), auditService, time.Now, func(now time.Time) (string, error) {
		id, idErr := identifier.NewUUIDv7(now, nil)
		return string(id), idErr
	}); err != nil {
		logger.Error("configure authorization management", "error", err)
		os.Exit(1)
	}
	organizationStore := organizationpostgres.New(pool)
	organizationService, err := organization.NewService(organizationStore, database.NewTransactor(pool), auditService, time.Now, func(now time.Time) (string, error) {
		id, idErr := identifier.NewUUIDv7(now, nil)
		return string(id), idErr
	})
	if err != nil {
		logger.Error("configure organization service", "error", err)
		os.Exit(1)
	}
	platformSigner, err := identity.NewAccessTokenSigner(cfg.Auth.Issuer, cfg.Auth.PlatformAudience, cfg.Auth.SigningKeyID, cfg.Auth.SigningKey, cfg.Auth.AccessLifetime)
	if err != nil {
		logger.Error("configure platform access tokens", "error", err)
		os.Exit(1)
	}
	platformVerifier := identity.NewAccessTokenVerifier(cfg.Auth.Issuer, cfg.Auth.PlatformAudience, map[string][]byte{cfg.Auth.SigningKeyID: cfg.Auth.SigningKey}, 5*time.Second)
	platformAdminService, err := platformadmin.NewService(platformadminpostgres.New(pool), platformSigner, platformVerifier, identity.DefaultPasswordParameters(), cfg.Auth.RefreshLifetime, time.Now, func(now time.Time) (string, error) {
		id, idErr := identifier.NewUUIDv7(now, nil)
		return string(id), idErr
	}, func() (string, error) { return identity.NewOpaqueToken(32) })
	if err != nil {
		logger.Error("configure platform administrator service", "error", err)
		os.Exit(1)
	}
	platformTenantService, err := platformtenant.NewService(platformtenantpostgres.New(pool), time.Now, func(now time.Time) (string, error) {
		id, idErr := identifier.NewUUIDv7(now, nil)
		return string(id), idErr
	})
	if err != nil {
		logger.Error("configure platform tenant service", "error", err)
		os.Exit(1)
	}
	provisioningService, err := provisioning.NewService(pool, identityService, organizationService, authorizationService, time.Now, func(now time.Time) (string, error) {
		id, idErr := identifier.NewUUIDv7(now, nil)
		return string(id), idErr
	}, func() (string, error) { return identity.NewOpaqueToken(32) }, cfg.Auth.InvitationLifetime)
	if err != nil {
		logger.Error("configure tenant provisioning service", "error", err)
		os.Exit(1)
	}
	server := httpserver.New(cfg.HTTP, logger, httpserver.Dependencies{Identity: identityService, Authorizer: authorizationService, Authorization: authorizationService, Organization: organizationService, PlatformAdmin: platformAdminService, PlatformTenant: platformTenantService, Provisioning: provisioningService, Ready: pool.Ping})

	errCh := make(chan error, 1)
	go func() {
		logger.Info("http server starting", "address", server.Addr)
		errCh <- server.ListenAndServe()
	}()

	select {
	case err := <-errCh:
		if !errors.Is(err, http.ErrServerClosed) {
			logger.Error("http server stopped unexpectedly", "error", err)
			os.Exit(1)
		}
	case <-ctx.Done():
		shutdownCtx, cancel := context.WithTimeout(context.Background(), cfg.HTTP.ShutdownTimeout)
		defer cancel()
		if err := server.Shutdown(shutdownCtx); err != nil {
			logger.Error("http server shutdown", "error", err)
			os.Exit(1)
		}
	}

	logger.Info("http server stopped")
}
