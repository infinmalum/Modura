// Package config loads and validates process configuration.
package config

import (
	"fmt"
	"os"
	"strconv"
	"time"
)

const (
	defaultAddress         = ":8080"
	defaultReadTimeout     = 10 * time.Second
	defaultWriteTimeout    = 15 * time.Second
	defaultIdleTimeout     = 60 * time.Second
	defaultShutdownTimeout = 10 * time.Second
	defaultMaxHeaderBytes  = 1 << 20
)

// Config contains the validated application configuration.
type Config struct{ HTTP HTTP }

// HTTP contains HTTP server configuration.
type HTTP struct {
	Address         string
	ReadTimeout     time.Duration
	WriteTimeout    time.Duration
	IdleTimeout     time.Duration
	ShutdownTimeout time.Duration
	MaxHeaderBytes  int
}

// FromEnv loads configuration from environment variables and applies defaults.
func FromEnv() (Config, error) {
	httpConfig := HTTP{Address: envOrDefault("MODURA_HTTP_ADDRESS", defaultAddress)}
	var err error
	if httpConfig.ReadTimeout, err = duration("MODURA_HTTP_READ_TIMEOUT", defaultReadTimeout); err != nil {
		return Config{}, err
	}
	if httpConfig.WriteTimeout, err = duration("MODURA_HTTP_WRITE_TIMEOUT", defaultWriteTimeout); err != nil {
		return Config{}, err
	}
	if httpConfig.IdleTimeout, err = duration("MODURA_HTTP_IDLE_TIMEOUT", defaultIdleTimeout); err != nil {
		return Config{}, err
	}
	if httpConfig.ShutdownTimeout, err = duration("MODURA_HTTP_SHUTDOWN_TIMEOUT", defaultShutdownTimeout); err != nil {
		return Config{}, err
	}
	if httpConfig.MaxHeaderBytes, err = integer("MODURA_HTTP_MAX_HEADER_BYTES", defaultMaxHeaderBytes); err != nil {
		return Config{}, err
	}
	return Config{HTTP: httpConfig}, nil
}

func envOrDefault(name, fallback string) string {
	if value := os.Getenv(name); value != "" {
		return value
	}
	return fallback
}

func duration(name string, fallback time.Duration) (time.Duration, error) {
	value := os.Getenv(name)
	if value == "" {
		return fallback, nil
	}
	parsed, err := time.ParseDuration(value)
	if err != nil {
		return 0, fmt.Errorf("parse %s: %w", name, err)
	}
	if parsed <= 0 {
		return 0, fmt.Errorf("%s must be positive", name)
	}
	return parsed, nil
}

func integer(name string, fallback int) (int, error) {
	value := os.Getenv(name)
	if value == "" {
		return fallback, nil
	}
	parsed, err := strconv.Atoi(value)
	if err != nil {
		return 0, fmt.Errorf("parse %s: %w", name, err)
	}
	if parsed <= 0 {
		return 0, fmt.Errorf("%s must be positive", name)
	}
	return parsed, nil
}
