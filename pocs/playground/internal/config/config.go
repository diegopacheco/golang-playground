package config

import (
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"time"
)

type Config struct {
	Addr       string
	GoBin      string
	GoRoot     string
	RunTimeout time.Duration
	MaxOutput  int
}

func Load() (Config, error) {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8086"
	}
	if _, err := strconv.Atoi(port); err != nil {
		return Config{}, fmt.Errorf("PORT %q: %w", port, err)
	}
	goBin := os.Getenv("GO_BIN")
	if goBin == "" {
		found, err := exec.LookPath("go")
		if err != nil {
			return Config{}, fmt.Errorf("go toolchain not found in PATH: %w", err)
		}
		goBin = found
	}
	out, err := exec.Command(goBin, "env", "GOROOT").Output()
	if err != nil {
		return Config{}, fmt.Errorf("go env GOROOT: %w", err)
	}
	return Config{
		Addr:       "127.0.0.1:" + port,
		GoBin:      goBin,
		GoRoot:     strings.TrimSpace(string(out)),
		RunTimeout: 10 * time.Second,
		MaxOutput:  64 * 1024,
	}, nil
}
