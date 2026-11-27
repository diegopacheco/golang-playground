package fsm

import (
	"bytes"
	"errors"
	"io"
	"testing"

	"github.com/hashicorp/raft"
)

func apply(t *testing.T, s *Store, cmd Command) any {
	t.Helper()
	b, err := Encode(cmd)
	if err != nil {
		t.Fatal(err)
	}
	return s.Apply(&raft.Log{Index: 1, Data: b})
}

func TestApplyKeepsEveryReplicaOnTheSameValue(t *testing.T) {
	tests := []struct {
		name    string
		cmds    []Command
		key     string
		want    string
		present bool
	}{
		{"set makes the value readable on every node", []Command{{Op: OpSet, Key: "a", Value: "1"}}, "a", "1", true},
		{"later set wins so replicas converge on the last write", []Command{{Op: OpSet, Key: "a", Value: "1"}, {Op: OpSet, Key: "a", Value: "2"}}, "a", "2", true},
		{"delete removes the key so stale values never resurface", []Command{{Op: OpSet, Key: "a", Value: "1"}, {Op: OpDelete, Key: "a"}}, "a", "", false},
		{"delete of a missing key is harmless so retries are safe", []Command{{Op: OpDelete, Key: "a"}}, "a", "", false},
		{"peer entries never leak into the user key space", []Command{{Op: OpPeer, Key: "a", Value: "localhost:9101"}}, "a", "", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := New()
			for _, c := range tt.cmds {
				if res := apply(t, s, c); res != nil {
					t.Fatalf("apply %v: %v", c, res)
				}
			}
			got, ok := s.Get(tt.key)
			if got != tt.want || ok != tt.present {
				t.Fatalf("Get(%q) = %q,%v want %q,%v", tt.key, got, ok, tt.want, tt.present)
			}
		})
	}
}

func TestApplyRejectsUnknownOpsInsteadOfSilentlyDivergingReplicas(t *testing.T) {
	s := New()
	res := apply(t, s, Command{Op: "rename", Key: "a"})
	err, ok := res.(error)
	if !ok || !errors.Is(err, ErrUnknownOp) {
		t.Fatalf("want ErrUnknownOp, got %v", res)
	}
}

func TestApplyRejectsCorruptEntries(t *testing.T) {
	s := New()
	if _, ok := s.Apply(&raft.Log{Index: 7, Data: []byte("{")}).(error); !ok {
		t.Fatal("corrupt log entry must surface an error")
	}
}

type memSink struct {
	bytes.Buffer
	cancelled bool
}

func (m *memSink) ID() string    { return "mem" }
func (m *memSink) Cancel() error { m.cancelled = true; return nil }
func (m *memSink) Close() error  { return nil }

func TestSnapshotRestoreRoundTripLetsLaggingNodesCatchUp(t *testing.T) {
	src := New()
	apply(t, src, Command{Op: OpSet, Key: "a", Value: "1"})
	apply(t, src, Command{Op: OpSet, Key: "b", Value: "2"})
	apply(t, src, Command{Op: OpPeer, Key: "node1", Value: "localhost:9101"})

	snap, err := src.Snapshot()
	if err != nil {
		t.Fatal(err)
	}
	apply(t, src, Command{Op: OpSet, Key: "after", Value: "x"})

	sink := &memSink{}
	if err := snap.Persist(sink); err != nil {
		t.Fatal(err)
	}
	snap.Release()

	dst := New()
	apply(t, dst, Command{Op: OpSet, Key: "stale", Value: "old"})
	if err := dst.Restore(io.NopCloser(&sink.Buffer)); err != nil {
		t.Fatal(err)
	}

	tests := []struct {
		name    string
		key     string
		want    string
		present bool
	}{
		{"keys from the snapshot are restored", "a", "1", true},
		{"every key from the snapshot is restored", "b", "2", true},
		{"writes after the snapshot point are not included", "after", "", false},
		{"restore replaces local state instead of merging stale data", "stale", "", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, ok := dst.Get(tt.key)
			if got != tt.want || ok != tt.present {
				t.Fatalf("Get(%q) = %q,%v want %q,%v", tt.key, got, ok, tt.want, tt.present)
			}
		})
	}
	if addr, ok := dst.Peer("node1"); !ok || addr != "localhost:9101" {
		t.Fatalf("peer address must survive restore so followers can still redirect, got %q,%v", addr, ok)
	}
}

func TestRestoreRejectsGarbageSnapshots(t *testing.T) {
	s := New()
	if err := s.Restore(io.NopCloser(bytes.NewBufferString("nope"))); err == nil {
		t.Fatal("garbage snapshot must fail restore")
	}
}
