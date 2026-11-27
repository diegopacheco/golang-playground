package node

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/hashicorp/go-hclog"
	"github.com/hashicorp/raft"
)

type member struct {
	node  *Node
	trans *raft.InmemTransport
}

func fastConfig() *raft.Config {
	c := raft.DefaultConfig()
	c.HeartbeatTimeout = 100 * time.Millisecond
	c.ElectionTimeout = 100 * time.Millisecond
	c.LeaderLeaseTimeout = 50 * time.Millisecond
	c.CommitTimeout = 5 * time.Millisecond
	return c
}

func startCluster(t *testing.T) []*member {
	t.Helper()
	members := make([]*member, 3)
	for i := range members {
		_, trans := raft.NewInmemTransportWithTimeout(raft.ServerAddress(fmt.Sprintf("raft%d", i+1)), 500*time.Millisecond)
		members[i] = &member{trans: trans}
	}
	for _, a := range members {
		for _, b := range members {
			if a != b {
				a.trans.Connect(b.trans.LocalAddr(), b.trans)
			}
		}
	}
	for i, m := range members {
		store := raft.NewInmemStore()
		n, err := New(Options{
			ID:        fmt.Sprintf("node%d", i+1),
			HTTPAddr:  fmt.Sprintf("localhost:910%d", i+1),
			Bootstrap: i == 0,
			Raft:      fastConfig(),
			Logs:      store,
			Stable:    store,
			Snaps:     raft.NewInmemSnapshotStore(),
			Transport: m.trans,
			Logger:    hclog.NewNullLogger(),
		})
		if err != nil {
			t.Fatal(err)
		}
		m.node = n
		t.Cleanup(func() { _ = n.Shutdown() })
	}
	leader := waitLeader(t, members)
	for _, m := range members {
		if m.node == leader {
			continue
		}
		req := JoinRequest{ID: m.node.id, RaftAddr: string(m.trans.LocalAddr()), HTTPAddr: m.node.httpAddr}
		if err := leader.Join(context.Background(), req); err != nil {
			t.Fatalf("join %s: %v", req.ID, err)
		}
	}
	return members
}

func eventually(t *testing.T, what string, cond func() bool) {
	t.Helper()
	deadline := time.Now().Add(10 * time.Second)
	for time.Now().Before(deadline) {
		if cond() {
			return
		}
		time.Sleep(20 * time.Millisecond)
	}
	t.Fatalf("timed out waiting for %s", what)
}

func waitLeader(t *testing.T, members []*member) *Node {
	t.Helper()
	var leader *Node
	eventually(t, "a leader", func() bool {
		for _, m := range members {
			if m.node.IsLeader() {
				leader = m.node
				return true
			}
		}
		return false
	})
	return leader
}

func leaderOf(members []*member) *member {
	for _, m := range members {
		if m.node.IsLeader() {
			return m
		}
	}
	return nil
}

func TestClusterReplicatesWritesAndSurvivesLeaderLoss(t *testing.T) {
	members := startCluster(t)
	ctx := context.Background()
	old := leaderOf(members)
	if old == nil {
		t.Fatal("no leader after join")
	}

	tests := []struct {
		name string
		run  func(t *testing.T)
	}{
		{"a write on the leader becomes readable on every replica", func(t *testing.T) {
			if err := old.node.Set(ctx, "k1", "v1"); err != nil {
				t.Fatal(err)
			}
			for _, m := range members {
				eventually(t, m.node.id+" to see k1", func() bool { v, ok := m.node.Get("k1"); return ok && v == "v1" })
			}
		}},
		{"followers reject writes so clients are sent to the single writer", func(t *testing.T) {
			for _, m := range members {
				if m == old {
					continue
				}
				if err := m.node.Set(ctx, "k", "v"); err == nil {
					t.Fatalf("%s accepted a write as follower", m.node.id)
				}
				addr, ok := m.node.LeaderHTTP()
				if !ok || addr != old.node.httpAddr {
					t.Fatalf("%s points clients to %q,%v want %q", m.node.id, addr, ok, old.node.httpAddr)
				}
			}
		}},
		{"losing the leader elects a new one that keeps old data and accepts writes", func(t *testing.T) {
			old.trans.DisconnectAll()
			for _, m := range members {
				if m != old {
					m.trans.Disconnect(old.trans.LocalAddr())
				}
			}
			if err := old.node.Shutdown(); err != nil {
				t.Fatal(err)
			}
			var survivors []*member
			for _, m := range members {
				if m != old {
					survivors = append(survivors, m)
				}
			}
			next := waitLeader(t, survivors)
			if next == old.node {
				t.Fatal("dead node still leader")
			}
			if v, ok := next.Get("k1"); !ok || v != "v1" {
				t.Fatalf("committed write lost after failover: %q,%v", v, ok)
			}
			if err := next.Set(ctx, "k2", "v2"); err != nil {
				t.Fatalf("new leader must accept writes with a 2 of 3 quorum: %v", err)
			}
			for _, m := range survivors {
				eventually(t, m.node.id+" to see k2", func() bool { v, ok := m.node.Get("k2"); return ok && v == "v2" })
				eventually(t, m.node.id+" to redirect to the new leader", func() bool {
					addr, ok := m.node.LeaderHTTP()
					return ok && addr == next.httpAddr
				})
			}
		}},
	}
	for _, tt := range tests {
		t.Run(tt.name, tt.run)
	}
}
