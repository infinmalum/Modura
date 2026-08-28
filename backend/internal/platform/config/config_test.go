package config

import (
	"testing"
	"time"
)

func TestFromEnvUsesDefaults(t *testing.T) {
	for _, name := range []string{"MODURA_HTTP_ADDRESS", "MODURA_HTTP_READ_TIMEOUT", "MODURA_HTTP_WRITE_TIMEOUT", "MODURA_HTTP_IDLE_TIMEOUT", "MODURA_HTTP_SHUTDOWN_TIMEOUT", "MODURA_HTTP_MAX_HEADER_BYTES"} {
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
	t.Setenv("MODURA_HTTP_READ_TIMEOUT", "never")
	if _, err := FromEnv(); err == nil {
		t.Fatal("FromEnv() error = nil, want an error")
	}
}

func TestFromEnvReadsValues(t *testing.T) {
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
