package dht

import (
	"crypto/sha1"
	"encoding/hex"
	"errors"
	"log"
	"sync"
	"time"

	"gitlab.lrz.de/netintum/teaching/p2psec_projects_2024/DHT-14/pkg/util"
)

// DHT represents a Distributed Hash Table.
type DHT struct {
	RoutingTable *RoutingTable
	Storage      *DHTStorage
}

// NewDHT creates a new instance of DHT.
func NewDHT(ttl time.Duration, encryptionKey []byte, nodeID string) *DHT {
	return &DHT{
		RoutingTable: NewRoutingTable(nodeID),
		Storage:      NewDHTStorage(ttl, encryptionKey),
	}
}

// PUT stores a value in the DHT.
func (d *DHT) PUT(key, value string, ttl int) error {
	return d.Storage.Put(key, value, ttl)
}

// GET retrieves a value from the DHT.
func (d *DHT) GET(key string) (string, error) {
	return d.Storage.Get(key)
}

func (d *DHT) FindValue(originID string, targetKeyID string) (string, []*KNode, error) {
	value, err := d.GET(targetKeyID)
	if err != nil {
		return "", nil, err
	}

	// Found value
	if value != "" {
		return value, nil, nil
	}

	// If the value is not found in the closest nodes, perform a FindNode RPC
	nodes := d.RoutingTable.GetClosestNodes(originID, targetKeyID)
	if len(nodes) == 0 {
		return "", nil, errors.New("no nodes found")
	}

	return "", nodes, nil
}

func (d *DHT) FindNode(originID string, targetID string) ([]*KNode, error) {
	nodes := d.RoutingTable.GetClosestNodes(originID, targetID)
	if len(nodes) == 0 {
		return nil, errors.New("no nodes found")
	}

	return nodes, nil
}

// TODO we dont need join leave here?
// Join allows the node to join the DHT network.
func (d *DHT) Join() {
	// Mock implementation
}

// Leave gracefully leaves the DHT network.
func (d *DHT) Leave() error {
	// Mock implementation
	return nil
}

// DHTStorage is a simple in-memory key-value store with TTL and encryption.
type DHTStorage struct {
	data map[string]*storageItem
	mu   sync.Mutex
	ttl  time.Duration
	key  []byte // Encryption key
}

type storageItem struct {
	value  string
	expiry time.Time
	hash   string
}

// NewDHTStorage initializes a new DHTStorage with a given TTL and encryption key.
func NewDHTStorage(ttl time.Duration, key []byte) *DHTStorage {
	storage := &DHTStorage{
		data: make(map[string]*storageItem),
		ttl:  ttl,
		key:  key,
	}
	storage.StartCleanup(ttl)
	return storage
}

// Put stores a value with the specified TTL.
func (s *DHTStorage) Put(key, value string, ttl int) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	encryptedValue, err := util.Encrypt([]byte(value), s.key)
	if err != nil {
		log.Printf("Error encrypting value: %v", err)
		return err
	}

	hasher := sha1.New()
	hasher.Write([]byte(encryptedValue))
	hash := hex.EncodeToString(hasher.Sum(nil))

	expiry := time.Now().Add(time.Duration(ttl) * time.Second)
	s.data[key] = &storageItem{
		value:  encryptedValue,
		expiry: expiry,
		hash:   hash,
	}
	return nil
}

// Get retrieves the value associated with the key if it exists and hasn't expired.
func (s *DHTStorage) Get(key string) (string, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	item, exists := s.data[key]
	if !exists {
		return "", nil
	}

	if time.Now().After(item.expiry) {
		delete(s.data, key)
		return "", nil
	}

	hasher := sha1.New()
	hasher.Write([]byte(item.value))
	hash := hex.EncodeToString(hasher.Sum(nil))
	if hash != item.hash {
		return "", errors.New("data integrity check failed")
	}

	decryptedValue, err := util.Decrypt(item.value, s.key)
	if err != nil {
		return "", err
	}

	return string(decryptedValue), nil
}

// StartCleanup starts a routine that periodically cleans up expired items.
func (s *DHTStorage) StartCleanup(interval time.Duration) {
	ticker := time.NewTicker(interval)
	go func() {
		for range ticker.C {
			s.CleanupExpired()
		}
	}()
}

// CleanupExpired removes expired items from the storage.
func (s *DHTStorage) CleanupExpired() {
	s.mu.Lock()
	defer s.mu.Unlock()

	for key, item := range s.data {
		if time.Now().After(item.expiry) {
			delete(s.data, key)
		}
	}
}
