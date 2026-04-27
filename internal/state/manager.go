package state

import (
	"context"
	"fmt"
	"sync"
	"time"
)

type Tunnel struct {
	ID        string
	ExpiresAt time.Time
	IsForever bool
	Cancel    context.CancelFunc
	// We might need to store the public port if we want to show it in `list`
	PublicPort int
}

type Manager struct {
	tunnels map[string]*Tunnel
	mu      sync.RWMutex
}

func NewManager() *Manager {
	m := &Manager{
		tunnels: make(map[string]*Tunnel),
	}
	go m.Reaper()
	return m
}

func (m *Manager) Add(id string, ttl time.Duration, cancel context.CancelFunc, publicPort int) *Tunnel {
	m.mu.Lock()
	defer m.mu.Unlock()

	expiresAt := time.Now().Add(ttl)
	isForever := ttl <= 0

	t := &Tunnel{
		ID:         id,
		ExpiresAt:  expiresAt,
		IsForever:  isForever,
		Cancel:     cancel,
		PublicPort: publicPort,
	}
	m.tunnels[id] = t
	return t
}

func (m *Manager) Remove(id string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.tunnels, id)
}

func (m *Manager) List() []*Tunnel {
	m.mu.RLock()
	defer m.mu.RUnlock()

	var list []*Tunnel
	for _, t := range m.tunnels {
		// Make a copy so we don't return mutable pointers to internal state, 
		// though returning pointer is okay for simple CLI.
		list = append(list, &Tunnel{
			ID:         t.ID,
			ExpiresAt:  t.ExpiresAt,
			IsForever:  t.IsForever,
			PublicPort: t.PublicPort,
		})
	}
	return list
}

func (m *Manager) Extend(id string, duration time.Duration) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	t, exists := m.tunnels[id]
	if !exists {
		return fmt.Errorf("tunnel %s not found", id)
	}

	if t.IsForever {
		return fmt.Errorf("tunnel %s is persistent and cannot be extended", id)
	}

	t.ExpiresAt = t.ExpiresAt.Add(duration)
	return nil
}

func (m *Manager) Persist(id string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	t, exists := m.tunnels[id]
	if !exists {
		return fmt.Errorf("tunnel %s not found", id)
	}

	t.IsForever = true
	return nil
}

// Reaper is a background goroutine that loops every 1 second, checks ExpiresAt, and calls the cancel function if time is up.
func (m *Manager) Reaper() {
	ticker := time.NewTicker(1 * time.Second)
	for range ticker.C {
		m.mu.Lock()
		now := time.Now()
		for id, t := range m.tunnels {
			if !t.IsForever && now.After(t.ExpiresAt) {
				fmt.Printf("Tunnel %s expired. Reaping...\n", id)
				t.Cancel()
				delete(m.tunnels, id)
			}
		}
		m.mu.Unlock()
	}
}
