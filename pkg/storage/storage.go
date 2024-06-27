package storage

import (
	"crypto/sha1"
	"encoding/hex"
	"sync"
	"time"
)

type Storage struct {
	data map[string]*storageItem
	mu   sync.Mutex
	ttl  time.Duration
}

type storageItem struct {
	value   string
	expiry  time.Time
	hash    string
}

func NewStorage(ttl time.Duration) *Storage {
	return &Storage{
		data: make(map[string]*storageItem),
		ttl:  ttl,
	}
}

func (s *Storage) Put(key, value string, ttl int) {
	s.mu.Lock()
	defer s.mu.Unlock()

	hasher := sha1.New()
	hasher.Write([]byte(value))
	hash := hex.EncodeToString(hasher.Sum(nil))

	expiry := time.Now().Add(time.Duration(ttl) * time.Second)
	s.data[key] = &storageItem{
		value:   value,
		expiry:  expiry,
		hash:    hash,
	}
}

func (s *Storage) Get(key string) string {
	s.mu.Lock()
	defer s.mu.Unlock()

	item, exists := s.data[key]
	if !exists {
		return ""
	}

	if time.Now().After(item.expiry) {
		delete(s.data, key)
		return ""
	}

	hasher := sha1.New()
	hasher.Write([]byte(item.value))
	hash := hex.EncodeToString(hasher.Sum(nil))
	if hash != item.hash {
		return ""
	}

	return item.value
}

func (s *Storage) cleanupExpired() {
	s.mu.Lock()
	defer s.mu.Unlock()

	for key, item := range s.data {
		if time.Now().After(item.expiry) {
			delete(s.data, key)
		}
	}
}

func (s *Storage) StartCleanup(interval time.Duration) {
	ticker := time.NewTicker(interval)
	go func() {
		for {
			<-ticker.C
			s.cleanupExpired()
		}
	}()
}