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
	"github.com/modura-dev/modura/backend/internal/modules/identity"
	identitypostgres "github.com/modura-dev/modura/backend/internal/modules/identity/postgres"
	"github.com/modura-dev/modura/backend/internal/platform/config"
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
	server := httpserver.New(cfg.HTTP, logger, httpserver.Dependencies{Identity: identityService, Ready: pool.Ping})

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
