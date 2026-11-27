package config

import (
	"errors"
	"flag"
	"fmt"
	"path/filepath"
)

var ErrInvalid = errors.New("invalid config")

type Config struct {
	NodeID    string
	HTTPAddr  string
	RaftAddr  string
	DataDir   string
	Bootstrap bool
	Join      string
	LogLevel  string
}

func Parse(args []string) (Config, error) {
	var c Config
	fs := flag.NewFlagSet("raftkv", flag.ContinueOnError)
	fs.StringVar(&c.NodeID, "id", "node1", "unique node id")
	fs.StringVar(&c.HTTPAddr, "http", "localhost:9101", "http listen and advertise address")
	fs.StringVar(&c.RaftAddr, "raft", "localhost:9201", "raft listen and advertise address")
	fs.StringVar(&c.DataDir, "data", "", "data directory, defaults to .run/data/<id>")
	fs.BoolVar(&c.Bootstrap, "bootstrap", false, "bootstrap a new cluster with this node")
	fs.StringVar(&c.Join, "join", "", "http address of a cluster member to join")
	fs.StringVar(&c.LogLevel, "log-level", "info", "debug, info, warn or error")
	if err := fs.Parse(args); err != nil {
		return Config{}, fmt.Errorf("parse flags: %w", err)
	}
	if c.DataDir == "" {
		c.DataDir = filepath.Join(".run", "data", c.NodeID)
	}
	if c.NodeID == "" || c.HTTPAddr == "" || c.RaftAddr == "" {
		return Config{}, fmt.Errorf("%w: id, http and raft are required", ErrInvalid)
	}
	if c.Bootstrap && c.Join != "" {
		return Config{}, fmt.Errorf("%w: bootstrap and join are mutually exclusive", ErrInvalid)
	}
	return c, nil
}
