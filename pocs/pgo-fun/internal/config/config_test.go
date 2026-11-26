package config

import (
	"errors"
	"log/slog"
	"testing"
)

func TestLoad(t *testing.T) {
	tests := []struct {
		name    string
		env     map[string]string
		addr    string
		level   slog.Level
		wantErr bool
	}{
		{name: "defaults bind the documented port on loopback so scripts and README agree", env: map[string]string{}, addr: "127.0.0.1:8082", level: slog.LevelInfo},
		{name: "scripts override the port from ports.env", env: map[string]string{"SERVICE_PORT": "9100"}, addr: "127.0.0.1:9100", level: slog.LevelInfo},
		{name: "debug level turns on per request logs when diagnosing", env: map[string]string{"LOG_LEVEL": "debug"}, addr: "127.0.0.1:8082", level: slog.LevelDebug},
		{name: "a non numeric port fails at startup instead of binding somewhere unexpected", env: map[string]string{"SERVICE_PORT": "abc"}, wantErr: true},
		{name: "an out of range port fails at startup", env: map[string]string{"SERVICE_PORT": "70000"}, wantErr: true},
		{name: "an unknown log level fails at startup", env: map[string]string{"LOG_LEVEL": "loud"}, wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg, err := Load(func(k string) string { return tt.env[k] })
			if tt.wantErr {
				if !errors.Is(err, ErrInvalid) {
					t.Fatalf("want ErrInvalid, got %v", err)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if cfg.Addr() != tt.addr || cfg.LogLevel != tt.level {
				t.Fatalf("got addr %s level %v", cfg.Addr(), cfg.LogLevel)
			}
		})
	}
}
