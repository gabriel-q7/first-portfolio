package cache

import (
	"context"
	"path"
	"sync"
	"time"

	"github.com/gabriel-q7/portfolio/backend/internal/domain/repository"
)

type entry struct {
	value     []byte
	expiresAt time.Time
}

// Memory is a small, process-local cache. Expired entries are removed
// opportunistically, avoiding a background cleanup goroutine.
type Memory struct {
	mu         sync.RWMutex
	items      map[string]entry
	maxEntries int
}

func New(maxEntries int) repository.CacheRepository {
	if maxEntries < 1 {
		maxEntries = 1
	}
	return &Memory{items: make(map[string]entry), maxEntries: maxEntries}
}

func (m *Memory) Get(_ context.Context, key string) ([]byte, error) {
	m.mu.RLock()
	item, ok := m.items[key]
	m.mu.RUnlock()
	if !ok {
		return nil, nil
	}
	if !item.expiresAt.IsZero() && time.Now().After(item.expiresAt) {
		m.mu.Lock()
		delete(m.items, key)
		m.mu.Unlock()
		return nil, nil
	}
	return append([]byte(nil), item.value...), nil
}

func (m *Memory) Set(_ context.Context, key string, value []byte, ttl time.Duration) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if len(m.items) >= m.maxEntries {
		m.removeExpiredLocked()
	}
	if len(m.items) >= m.maxEntries {
		for existingKey := range m.items {
			delete(m.items, existingKey)
			break
		}
	}
	item := entry{value: append([]byte(nil), value...)}
	if ttl > 0 {
		item.expiresAt = time.Now().Add(ttl)
	}
	m.items[key] = item
	return nil
}

func (m *Memory) Delete(_ context.Context, key string) error {
	m.mu.Lock()
	delete(m.items, key)
	m.mu.Unlock()
	return nil
}

func (m *Memory) Exists(ctx context.Context, key string) (bool, error) {
	value, err := m.Get(ctx, key)
	return value != nil, err
}

func (m *Memory) FlushPattern(_ context.Context, pattern string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	for key := range m.items {
		if matched, _ := path.Match(pattern, key); matched {
			delete(m.items, key)
		}
	}
	return nil
}

func (m *Memory) removeExpiredLocked() {
	now := time.Now()
	for key, item := range m.items {
		if !item.expiresAt.IsZero() && now.After(item.expiresAt) {
			delete(m.items, key)
		}
	}
}
