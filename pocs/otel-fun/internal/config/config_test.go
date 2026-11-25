package config

import (
	"testing"
	"time"
)

func TestLoad(t *testing.T) {
	tests := []struct {
		name    string
		env     map[string]string
		wantErr bool
		check   func(t *testing.T, c Config)
	}{
		{
			name: "defaults match scripts ports so the service runs with no env at all",
			env:  map[string]string{},
			check: func(t *testing.T, c Config) {
				if c.Port != 8083 || c.OTLPEndpoint != "localhost:4318" || !c.ConsoleExport || c.ServiceName != "otel-fun" {
					t.Fatalf("unexpected defaults %+v", c)
				}
			},
		},
		{
			name: "env overrides let scripts move ports without code changes",
			env:  map[string]string{"PORT": "9000", "OTLP_ENDPOINT": "jaeger:4318", "CONSOLE_EXPORT": "false", "METRIC_INTERVAL": "2s"},
			check: func(t *testing.T, c Config) {
				if c.Port != 9000 || c.OTLPEndpoint != "jaeger:4318" || c.ConsoleExport || c.MetricInterval != 2*time.Second {
					t.Fatalf("overrides not applied %+v", c)
				}
				if c.Addr() != ":9000" {
					t.Fatalf("addr %q", c.Addr())
				}
			},
		},
		{name: "non numeric port fails fast instead of binding a random port", env: map[string]string{"PORT": "abc"}, wantErr: true},
		{name: "out of range port fails fast", env: map[string]string{"PORT": "70000"}, wantErr: true},
		{name: "bad console flag fails fast instead of silently disabling export", env: map[string]string{"CONSOLE_EXPORT": "maybe"}, wantErr: true},
		{name: "bad metric interval fails fast", env: map[string]string{"METRIC_INTERVAL": "soon"}, wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c, err := Load(func(k string) string { return tt.env[k] })
			if tt.wantErr {
				if err == nil {
					t.Fatal("expected error")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error %v", err)
			}
			tt.check(t, c)
		})
	}
}
