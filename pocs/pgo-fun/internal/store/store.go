package store

import (
	"sync"

	"github.com/diegopacheco/golang-playground/pocs/pgo-fun/internal/order"
)

type Memory struct {
	mu     sync.RWMutex
	orders map[string]order.Priced
}

func NewMemory() *Memory {
	return &Memory{orders: make(map[string]order.Priced)}
}

func (m *Memory) Put(p order.Priced) {
	m.mu.Lock()
	m.orders[p.ID] = p
	m.mu.Unlock()
}

func (m *Memory) Get(id string) (order.Priced, bool) {
	m.mu.RLock()
	p, ok := m.orders[id]
	m.mu.RUnlock()
	return p, ok
}
