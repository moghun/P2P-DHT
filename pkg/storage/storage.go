package storage

import (
    "sync"
)

type Storage struct {
    mu    sync.RWMutex
    store map[string][]byte
}

func NewStorage() *Storage {
    return &Storage{
        store: make(map[string][]byte),
    }
}

func (s *Storage) Put(key string, value []byte) {
    s.mu.Lock()
    defer s.mu.Unlock()
    s.store[key] = value
}

func (s *Storage) Get(key string) ([]byte, bool) {
    s.mu.RLock()
    defer s.mu.RUnlock()
    value, exists := s.store[key]
    return value, exists
}

func (s *Storage) Delete(key string) {
    s.mu.Lock()
    defer s.mu.Unlock()
    delete(s.store, key)
}
