package node

import (
	"fmt"
	"io"
	"net"
	"os"
	"path/filepath"
	"time"

	"github.com/hashicorp/go-hclog"
	"github.com/hashicorp/raft"
	raftboltdb "github.com/hashicorp/raft-boltdb/v2"
)

type DiskConfig struct {
	ID        string
	HTTPAddr  string
	RaftAddr  string
	DataDir   string
	Bootstrap bool
	Logger    hclog.Logger
}

func Open(c DiskConfig) (*Node, error) {
	if err := os.MkdirAll(c.DataDir, 0o755); err != nil {
		return nil, fmt.Errorf("create data dir: %w", err)
	}
	bolt, err := raftboltdb.NewBoltStore(filepath.Join(c.DataDir, "raft.db"))
	if err != nil {
		return nil, fmt.Errorf("open bolt store: %w", err)
	}
	snaps, err := raft.NewFileSnapshotStoreWithLogger(c.DataDir, 2, c.Logger)
	if err != nil {
		bolt.Close()
		return nil, fmt.Errorf("open snapshot store: %w", err)
	}
	advertise, err := net.ResolveTCPAddr("tcp", c.RaftAddr)
	if err != nil {
		bolt.Close()
		return nil, fmt.Errorf("resolve raft addr: %w", err)
	}
	trans, err := raft.NewTCPTransportWithLogger(advertise.String(), advertise, 3, 10*time.Second, c.Logger)
	if err != nil {
		bolt.Close()
		return nil, fmt.Errorf("open raft transport: %w", err)
	}
	n, err := New(Options{
		ID:        c.ID,
		HTTPAddr:  c.HTTPAddr,
		Bootstrap: c.Bootstrap,
		Logs:      bolt,
		Stable:    bolt,
		Snaps:     snaps,
		Transport: trans,
		Logger:    c.Logger,
		Closers:   []io.Closer{bolt},
	})
	if err != nil {
		trans.Close()
		bolt.Close()
		return nil, err
	}
	return n, nil
}
