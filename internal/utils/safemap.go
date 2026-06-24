package utils

import (
	"sync"
)

type SafeMap[T any] struct {
	mu sync.Mutex
	data map[string]T
}

func NewSafeMap[T any]() *SafeMap[T] {
    return &SafeMap[T]{
        data: make(map[string]T),
    }
}

func (m *SafeMap[T]) Set(key string, value T) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.data[key] = value
}

func (m *SafeMap[T]) Get(key string) (T, bool) {
	m.mu.Lock()
	defer m.mu.Unlock()
	val, exists := m.data[key]
	return val, exists
}