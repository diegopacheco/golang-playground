package config

import (
	"fmt"
	"strconv"
	"time"
)

type Config struct {
	ServiceName     string
	Port            int
	OTLPEndpoint    string
	ConsoleExport   bool
	MetricInterval  time.Duration
	ShutdownTimeout time.Duration
}

func Load(getenv func(string) string) (Config, error) {
	cfg := Config{
		ServiceName:     "otel-fun",
		Port:            8083,
		OTLPEndpoint:    "localhost:4318",
		ConsoleExport:   true,
		MetricInterval:  15 * time.Second,
		ShutdownTimeout: 10 * time.Second,
	}
	if v := getenv("SERVICE_NAME"); v != "" {
		cfg.ServiceName = v
	}
	if v := getenv("PORT"); v != "" {
		port, err := strconv.Atoi(v)
		if err != nil || port < 1 || port > 65535 {
			return Config{}, fmt.Errorf("invalid PORT %q", v)
		}
		cfg.Port = port
	}
	if v := getenv("OTLP_ENDPOINT"); v != "" {
		cfg.OTLPEndpoint = v
	}
	if v := getenv("CONSOLE_EXPORT"); v != "" {
		enabled, err := strconv.ParseBool(v)
		if err != nil {
			return Config{}, fmt.Errorf("invalid CONSOLE_EXPORT %q: %w", v, err)
		}
		cfg.ConsoleExport = enabled
	}
	if v := getenv("METRIC_INTERVAL"); v != "" {
		interval, err := time.ParseDuration(v)
		if err != nil {
			return Config{}, fmt.Errorf("invalid METRIC_INTERVAL %q: %w", v, err)
		}
		cfg.MetricInterval = interval
	}
	return cfg, nil
}

func (c Config) Addr() string {
	return fmt.Sprintf(":%d", c.Port)
}
