package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"
)

type memoryItem struct {
	data      []byte
	expiresAt time.Time
}

type MemoryCache struct {
	mu   sync.RWMutex
	data map[string]memoryItem
	ttl  time.Duration
}

func NewMemoryCache(ttl int) *MemoryCache {
	return &MemoryCache{
		data: make(map[string]memoryItem),
		ttl:  time.Duration(ttl) * time.Second,
	}
}

func (m *MemoryCache) Set(ctx context.Context, key string, value interface{}) error {
	data, err := json.Marshal(value)
	if err != nil {
		return fmt.Errorf("marshal value: %w", err)
	}
	m.mu.Lock()
	m.data[key] = memoryItem{
		data:      data,
		expiresAt: time.Now().Add(m.ttl),
	}
	m.mu.Unlock()
	return nil
}

func (m *MemoryCache) SetWithTTL(ctx context.Context, key string, value interface{}, ttl time.Duration) error {
	data, err := json.Marshal(value)
	if err != nil {
		return fmt.Errorf("marshal value: %w", err)
	}
	m.mu.Lock()
	m.data[key] = memoryItem{
		data:      data,
		expiresAt: time.Now().Add(ttl),
	}
	m.mu.Unlock()
	return nil
}

func (m *MemoryCache) Get(ctx context.Context, key string, dest interface{}) error {
	m.mu.RLock()
	item, ok := m.data[key]
	m.mu.RUnlock()

	if !ok {
		return fmt.Errorf("cache miss: key %s not found", key)
	}

	if time.Now().After(item.expiresAt) {
		m.mu.Lock()
		delete(m.data, key)
		m.mu.Unlock()
		return fmt.Errorf("cache miss: key %s expired", key)
	}

	return json.Unmarshal(item.data, dest)
}

func (m *MemoryCache) Delete(ctx context.Context, keys ...string) error {
	m.mu.Lock()
	for _, key := range keys {
		delete(m.data, key)
	}
	m.mu.Unlock()
	return nil
}

func (m *MemoryCache) FlushAll(ctx context.Context) error {
	m.mu.Lock()
	m.data = make(map[string]memoryItem)
	m.mu.Unlock()
	return nil
}
