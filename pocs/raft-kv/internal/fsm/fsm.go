package fsm

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"maps"
	"sync"

	"github.com/hashicorp/raft"
)

type Op string

const (
	OpSet    Op = "set"
	OpDelete Op = "delete"
	OpPeer   Op = "peer"
)

var ErrUnknownOp = errors.New("unknown op")

type Command struct {
	Op    Op     `json:"op"`
	Key   string `json:"key"`
	Value string `json:"value,omitempty"`
}

type State struct {
	Data  map[string]string `json:"data"`
	Peers map[string]string `json:"peers"`
}

type Store struct {
	mu    sync.RWMutex
	data  map[string]string
	peers map[string]string
}

func New() *Store {
	return &Store{data: map[string]string{}, peers: map[string]string{}}
}

func Encode(cmd Command) ([]byte, error) {
	b, err := json.Marshal(cmd)
	if err != nil {
		return nil, fmt.Errorf("encode command: %w", err)
	}
	return b, nil
}

func (s *Store) Apply(entry *raft.Log) any {
	var cmd Command
	if err := json.Unmarshal(entry.Data, &cmd); err != nil {
		return fmt.Errorf("decode command at index %d: %w", entry.Index, err)
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	switch cmd.Op {
	case OpSet:
		s.data[cmd.Key] = cmd.Value
	case OpDelete:
		delete(s.data, cmd.Key)
	case OpPeer:
		s.peers[cmd.Key] = cmd.Value
	default:
		return fmt.Errorf("%w: %q", ErrUnknownOp, cmd.Op)
	}
	return nil
}

func (s *Store) Get(key string) (string, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	v, ok := s.data[key]
	return v, ok
}

func (s *Store) Peer(id string) (string, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	v, ok := s.peers[id]
	return v, ok
}

func (s *Store) Len() int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return len(s.data)
}

func (s *Store) Snapshot() (raft.FSMSnapshot, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return &snapshot{state: State{Data: maps.Clone(s.data), Peers: maps.Clone(s.peers)}}, nil
}

func (s *Store) Restore(rc io.ReadCloser) error {
	defer rc.Close()
	var st State
	if err := json.NewDecoder(rc).Decode(&st); err != nil {
		return fmt.Errorf("restore snapshot: %w", err)
	}
	if st.Data == nil {
		st.Data = map[string]string{}
	}
	if st.Peers == nil {
		st.Peers = map[string]string{}
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	s.data = st.Data
	s.peers = st.Peers
	return nil
}

type snapshot struct {
	state State
}

func (s *snapshot) Persist(sink raft.SnapshotSink) error {
	if err := json.NewEncoder(sink).Encode(s.state); err != nil {
		_ = sink.Cancel()
		return fmt.Errorf("persist snapshot: %w", err)
	}
	if err := sink.Close(); err != nil {
		return fmt.Errorf("close snapshot sink: %w", err)
	}
	return nil
}

func (s *snapshot) Release() {}
