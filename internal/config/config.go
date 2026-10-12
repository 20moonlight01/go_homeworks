package config

import (
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
)

type Config struct {
	HTTPAddr        string
	LogLevel        string
	ShutdownTimeout time.Duration

	HTTPReadTimeout       time.Duration
	HTTPReadHeaderTimeout time.Duration
	HTTPIdleTimeout       time.Duration
	HTTPWriteTimeout      time.Duration

	DatabaseURL       string
	DBMaxConns        int32
	DBMinConns        int32
	DBMaxConnLifetime time.Duration
	DBConnectTimeout  time.Duration
	DBQueryTimeout    time.Duration
}

func Load() (*Config, error) {
	cfg := &Config{
		HTTPAddr:    os.Getenv("HTTP_ADDR"),
		DatabaseURL: os.Getenv("DATABASE_URL"),
	}

	if cfg.HTTPAddr == "" {
		return nil, errors.New("HTTP_ADDR is required")
	}
	if cfg.DatabaseURL == "" {
		return nil, errors.New("DATABASE_URL is required")
	}

	cfg.LogLevel = "info"
	if v := os.Getenv("LOG_LEVEL"); v != "" {
		cfg.LogLevel = v
	}

	var err error

	if cfg.DBMaxConns, err = getInt32("DATABASE_MAX_CONNS"); err != nil {
		return nil, err
	}
	if cfg.DBMinConns, err = getInt32("DATABASE_MIN_CONNS"); err != nil {
		return nil, err
	}
	if cfg.DBMinConns > cfg.DBMaxConns {
		return nil, errors.New("DATABASE_MIN_CONNS must not be greater than DATABASE_MAX_CONNS")
	}

	if cfg.DBMaxConnLifetime, err = getDuration("DATABASE_MAX_CONN_LIFETIME"); err != nil {
		return nil, err
	}
	if cfg.DBConnectTimeout, err = getDuration("DATABASE_CONNECT_TIMEOUT"); err != nil {
		return nil, err
	}
	if cfg.DBQueryTimeout, err = getDuration("DATABASE_QUERY_TIMEOUT"); err != nil {
		return nil, err
	}
	if cfg.ShutdownTimeout, err = getDuration("SHUTDOWN_TIMEOUT"); err != nil {
		return nil, err
	}

	if cfg.HTTPReadTimeout, err = getDurationDefault("HTTP_READ_TIMEOUT", 10*time.Second); err != nil {
		return nil, err
	}
	if cfg.HTTPReadHeaderTimeout, err = getDurationDefault("HTTP_READ_HEADER_TIMEOUT", 5*time.Second); err != nil {
		return nil, err
	}
	if cfg.HTTPIdleTimeout, err = getDurationDefault("HTTP_IDLE_TIMEOUT", 60*time.Second); err != nil {
		return nil, err
	}
	if cfg.HTTPWriteTimeout, err = getDurationDefault("HTTP_WRITE_TIMEOUT", 15*time.Second); err != nil {
		return nil, err
	}

	if err := validateLogLevel(cfg.LogLevel); err != nil {
		return nil, err
	}

	return cfg, nil
}

func getDuration(key string) (time.Duration, error) {
	v := strings.TrimSpace(os.Getenv(key))
	if v == "" {
		return 0, fmt.Errorf("%s is required", key)
	}
	d, err := time.ParseDuration(v)
	if err != nil {
		return 0, fmt.Errorf("%s: parse duration: %w", key, err)
	}
	if d <= 0 {
		return 0, fmt.Errorf("%s: must be positive", key)
	}
	return d, nil
}

func getDurationDefault(key string, def time.Duration) (time.Duration, error) {
	v := strings.TrimSpace(os.Getenv(key))
	if v == "" {
		return def, nil
	}
	d, err := time.ParseDuration(v)
	if err != nil {
		return 0, fmt.Errorf("%s: parse duration: %w", key, err)
	}
	if d <= 0 {
		return 0, fmt.Errorf("%s: must be positive", key)
	}
	return d, nil
}

func getInt32(key string) (int32, error) {
	v := strings.TrimSpace(os.Getenv(key))
	if v == "" {
		return 0, fmt.Errorf("%s is required", key)
	}
	n, err := strconv.ParseInt(v, 10, 32)
	if err != nil {
		return 0, fmt.Errorf("%s: parse int: %w", key, err)
	}
	if n <= 0 {
		return 0, fmt.Errorf("%s: must be positive", key)
	}
	return int32(n), nil
}

func validateLogLevel(level string) error {
	switch strings.ToLower(strings.TrimSpace(level)) {
	case "debug", "info", "warn", "error":
		return nil
	default:
		return fmt.Errorf("LOG_LEVEL: unknown level %q", level)
	}
}
