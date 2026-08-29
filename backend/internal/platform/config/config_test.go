package config

import (
	"testing"
	"time"
)

func TestFromEnvUsesDefaults(t *testing.T) {
	setRequired(t)
	for _, name := range []string{"MODURA_HTTP_ADDRESS", "MODURA_HTTP_READ_TIMEOUT", "MODURA_HTTP_WRITE_TIMEOUT", "MODURA_HTTP_IDLE_TIMEOUT", "MODURA_HTTP_SHUTDOWN_TIMEOUT", "MODURA_HTTP_MAX_HEADER_BYTES", "MODURA_AUTH_COOKIE_SECURE", "MODURA_AUTH_ACCESS_LIFETIME", "MODURA_AUTH_REFRESH_LIFETIME"} {
		t.Setenv(name, "")
	}
	cfg, err := FromEnv()
	if err != nil {
		t.Fatalf("FromEnv() error = %v", err)
	}
	if cfg.HTTP.Address != defaultAddress {
		t.Fatalf("Address = %q, want %q", cfg.HTTP.Address, defaultAddress)
	}
	if cfg.HTTP.ReadTimeout != defaultReadTimeout {
		t.Fatalf("ReadTimeout = %v, want %v", cfg.HTTP.ReadTimeout, defaultReadTimeout)
	}
}

func TestFromEnvRejectsInvalidDuration(t *testing.T) {
	setRequired(t)
	t.Setenv("MODURA_HTTP_READ_TIMEOUT", "never")
	if _, err := FromEnv(); err == nil {
		t.Fatal("FromEnv() error = nil, want an error")
	}
}

func TestFromEnvReadsValues(t *testing.T) {
	setRequired(t)
	t.Setenv("MODURA_HTTP_ADDRESS", "127.0.0.1:9000")
	t.Setenv("MODURA_HTTP_READ_TIMEOUT", "3s")
	cfg, err := FromEnv()
	if err != nil {
		t.Fatalf("FromEnv() error = %v", err)
	}
	if cfg.HTTP.Address != "127.0.0.1:9000" || cfg.HTTP.ReadTimeout != 3*time.Second {
		t.Fatalf("unexpected config: %+v", cfg.HTTP)
	}
}

func TestFromEnvRequiresSecrets(t *testing.T) {
	t.Setenv("MODURA_DATABASE_URL", "")
	t.Setenv("MODURA_AUTH_SIGNING_KEY", "")
	if _, err := FromEnv(); err == nil {
		t.Fatal("FromEnv() error = nil, want an error")
	}
}

func setRequired(t *testing.T) {
	t.Helper()
	t.Setenv("MODURA_DATABASE_URL", "postgres://localhost/modura")
	t.Setenv("MODURA_AUTH_SIGNING_KEY", "test-only-signing-key-with-32-bytes")
}
