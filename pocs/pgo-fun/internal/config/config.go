package config

import (
	"errors"
	"fmt"
	"log/slog"
	"strconv"
	"time"
)

var ErrInvalid = errors.New("invalid config")

type Config struct {
	Host            string
	Port            int
	LogLevel        slog.Level
	ShutdownTimeout time.Duration
}

func (c Config) Addr() string {
	return fmt.Sprintf("%s:%d", c.Host, c.Port)
}

func Load(getenv func(string) string) (Config, error) {
	cfg := Config{Host: "127.0.0.1", Port: 8082, LogLevel: slog.LevelInfo, ShutdownTimeout: 10 * time.Second}
	if v := getenv("SERVICE_HOST"); v != "" {
		cfg.Host = v
	}
	if v := getenv("SERVICE_PORT"); v != "" {
		port, err := strconv.Atoi(v)
		if err != nil || port < 1 || port > 65535 {
			return Config{}, fmt.Errorf("%w: SERVICE_PORT %q", ErrInvalid, v)
		}
		cfg.Port = port
	}
	if v := getenv("LOG_LEVEL"); v != "" {
		if err := cfg.LogLevel.UnmarshalText([]byte(v)); err != nil {
			return Config{}, fmt.Errorf("%w: LOG_LEVEL %q: %w", ErrInvalid, v, err)
		}
	}
	return cfg, nil
}
