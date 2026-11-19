package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
)

type Config struct {
	Port     int
	DBPath   string
	Headless bool
}

func Load() (Config, error) {
	port := 8090
	if v := os.Getenv("BIZ_PORT"); v != "" {
		p, err := strconv.Atoi(v)
		if err != nil {
			return Config{}, fmt.Errorf("BIZ_PORT: %w", err)
		}
		port = p
	}
	dbPath := os.Getenv("BIZ_DB")
	if dbPath == "" {
		dir, err := os.UserConfigDir()
		if err != nil {
			return Config{}, fmt.Errorf("user config dir: %w", err)
		}
		dbPath = filepath.Join(dir, "wails-fun", "biz.db")
	}
	headless := false
	if v := os.Getenv("BIZ_HEADLESS"); v != "" {
		h, err := strconv.ParseBool(v)
		if err != nil {
			return Config{}, fmt.Errorf("BIZ_HEADLESS: %w", err)
		}
		headless = h
	}
	return Config{Port: port, DBPath: dbPath, Headless: headless}, nil
}
