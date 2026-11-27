package node

import (
	"context"
	"errors"
	"fmt"
	"io"
	"sync"
	"time"

	"github.com/diegopacheco/golang-playground/pocs/raft-kv/internal/fsm"
	"github.com/hashicorp/go-hclog"
	"github.com/hashicorp/raft"
)

const applyTimeout = 5 * time.Second

var (
	ErrNotLeader  = errors.New("not the leader")
	ErrBadRequest = errors.New("bad request")
)

type JoinRequest struct {
	ID       string `json:"id"`
	RaftAddr string `json:"raft_addr"`
	HTTPAddr string `json:"http_addr"`
}

type Status struct {
	ID             string `json:"id"`
	Role           string `json:"role"`
	LeaderID       string `json:"leader_id"`
	LeaderRaftAddr string `json:"leader_raft_addr"`
	LeaderHTTPAddr string `json:"leader_http_addr"`
	Term           uint64 `json:"term"`
	LastIndex      uint64 `json:"last_index"`
	AppliedIndex   uint64 `json:"applied_index"`
	Keys           int    `json:"keys"`
}

type Options struct {
	ID        string
	HTTPAddr  string
	Bootstrap bool
	Raft      *raft.Config
	Logs      raft.LogStore
	Stable    raft.StableStore
	Snaps     raft.SnapshotStore
	Transport raft.Transport
	Logger    hclog.Logger
	Closers   []io.Closer
}

type Node struct {
	id       string
	httpAddr string
	fresh    bool
	raft     *raft.Raft
	fsm      *fsm.Store
	closers  []io.Closer
	stop     chan struct{}
	done     chan struct{}
	once     sync.Once
}

func New(opts Options) (*Node, error) {
	conf := raft.DefaultConfig()
	if opts.Raft != nil {
		conf = opts.Raft
	}
	notify := make(chan bool, 8)
	conf.LocalID = raft.ServerID(opts.ID)
	conf.NotifyCh = notify
	if opts.Logger != nil {
		conf.Logger = opts.Logger
	}

	existing, err := raft.HasExistingState(opts.Logs, opts.Stable, opts.Snaps)
	if err != nil {
		return nil, fmt.Errorf("check existing state: %w", err)
	}
	if opts.Bootstrap && !existing {
		servers := raft.Configuration{Servers: []raft.Server{{ID: conf.LocalID, Address: opts.Transport.LocalAddr()}}}
		if err := raft.BootstrapCluster(conf, opts.Logs, opts.Stable, opts.Snaps, opts.Transport, servers); err != nil {
			return nil, fmt.Errorf("bootstrap cluster: %w", err)
		}
	}

	store := fsm.New()
	r, err := raft.NewRaft(conf, store, opts.Logs, opts.Stable, opts.Snaps, opts.Transport)
	if err != nil {
		return nil, fmt.Errorf("start raft: %w", err)
	}
	n := &Node{
		id:       opts.ID,
		httpAddr: opts.HTTPAddr,
		fresh:    !existing,
		raft:     r,
		fsm:      store,
		closers:  opts.Closers,
		stop:     make(chan struct{}),
		done:     make(chan struct{}),
	}
	go n.watchLeadership(notify)
	return n, nil
}

func (n *Node) watchLeadership(notify <-chan bool) {
	defer close(n.done)
	for {
		select {
		case <-n.stop:
			return
		case leader := <-notify:
			if leader {
				_ = n.apply(fsm.Command{Op: fsm.OpPeer, Key: n.id, Value: n.httpAddr}, applyTimeout)
			}
		}
	}
}

func (n *Node) Fresh() bool { return n.fresh }

func (n *Node) IsLeader() bool { return n.raft.State() == raft.Leader }

func (n *Node) LeaderHTTP() (string, bool) {
	_, id := n.raft.LeaderWithID()
	if id == "" {
		return "", false
	}
	return n.fsm.Peer(string(id))
}

func (n *Node) Get(key string) (string, bool) { return n.fsm.Get(key) }

func (n *Node) Set(ctx context.Context, key, value string) error {
	return n.applyCtx(ctx, fsm.Command{Op: fsm.OpSet, Key: key, Value: value})
}

func (n *Node) Delete(ctx context.Context, key string) error {
	return n.applyCtx(ctx, fsm.Command{Op: fsm.OpDelete, Key: key})
}

func (n *Node) Join(ctx context.Context, req JoinRequest) error {
	if req.ID == "" || req.RaftAddr == "" || req.HTTPAddr == "" {
		return fmt.Errorf("%w: id, raft_addr and http_addr are required", ErrBadRequest)
	}
	timeout := timeoutFrom(ctx)
	f := n.raft.AddVoter(raft.ServerID(req.ID), raft.ServerAddress(req.RaftAddr), 0, timeout)
	if err := f.Error(); err != nil {
		return wrapRaft("add voter "+req.ID, err)
	}
	return n.applyCtx(ctx, fsm.Command{Op: fsm.OpPeer, Key: req.ID, Value: req.HTTPAddr})
}

func (n *Node) Status() Status {
	addr, id := n.raft.LeaderWithID()
	leaderHTTP, _ := n.fsm.Peer(string(id))
	return Status{
		ID:             n.id,
		Role:           n.raft.State().String(),
		LeaderID:       string(id),
		LeaderRaftAddr: string(addr),
		LeaderHTTPAddr: leaderHTTP,
		Term:           n.raft.CurrentTerm(),
		LastIndex:      n.raft.LastIndex(),
		AppliedIndex:   n.raft.AppliedIndex(),
		Keys:           n.fsm.Len(),
	}
}

func (n *Node) Shutdown() error {
	var errs []error
	n.once.Do(func() {
		close(n.stop)
		<-n.done
		if err := n.raft.Shutdown().Error(); err != nil {
			errs = append(errs, fmt.Errorf("shutdown raft: %w", err))
		}
		for _, c := range n.closers {
			if err := c.Close(); err != nil {
				errs = append(errs, fmt.Errorf("close store: %w", err))
			}
		}
	})
	return errors.Join(errs...)
}

func (n *Node) applyCtx(ctx context.Context, cmd fsm.Command) error {
	if err := ctx.Err(); err != nil {
		return fmt.Errorf("apply %s: %w", cmd.Op, err)
	}
	return n.apply(cmd, timeoutFrom(ctx))
}

func (n *Node) apply(cmd fsm.Command, timeout time.Duration) error {
	b, err := fsm.Encode(cmd)
	if err != nil {
		return err
	}
	f := n.raft.Apply(b, timeout)
	if err := f.Error(); err != nil {
		return wrapRaft(string(cmd.Op), err)
	}
	if err, ok := f.Response().(error); ok {
		return fmt.Errorf("apply %s: %w", cmd.Op, err)
	}
	return nil
}

func wrapRaft(op string, err error) error {
	if errors.Is(err, raft.ErrNotLeader) || errors.Is(err, raft.ErrLeadershipLost) {
		return fmt.Errorf("%s: %w: %w", op, ErrNotLeader, err)
	}
	return fmt.Errorf("%s: %w", op, err)
}

func timeoutFrom(ctx context.Context) time.Duration {
	if d, ok := ctx.Deadline(); ok {
		return time.Until(d)
	}
	return applyTimeout
}
