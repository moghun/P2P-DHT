package storage

import (
	"crypto/sha1"
	"encoding/hex"
	"errors"
	"log"
	"sync"
	"time"

	"gitlab.lrz.de/netintum/teaching/p2psec_projects_2024/DHT-14/pkg/util"
)

type Storage struct {
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

func NewStorage(ttl time.Duration, key []byte) *Storage {
	return &Storage{
		data: make(map[string]*storageItem),
		ttl:  ttl,
		key:  key,
	}
}

func (s *Storage) Put(key, value string, ttl int) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	encryptedValue, err := util.Encrypt([]byte(value), s.key)
	if err != nil {
		log.Printf("Error encrypting value: %v", err)
		return err
	}
	log.Printf("Encrypted value: %s", encryptedValue)

	hasher := sha1.New()
	hasher.Write([]byte(encryptedValue))
	hash := hex.EncodeToString(hasher.Sum(nil))
	log.Printf("SHA1 hash of encrypted value: %s", hash)

	expiry := time.Now().Add(time.Duration(ttl) * time.Second)
	s.data[key] = &storageItem{
		value:  encryptedValue,
		expiry: expiry,
		hash:   hash,
	}
	log.Printf("Stored item: key=%s, value=%s, expiry=%s, hash=%s", key, encryptedValue, expiry.String(), hash)
	return nil
}

func (s *Storage) Get(key string) (string, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	item, exists := s.data[key]
	if !exists {
		log.Printf("Key not found: %s", key)
		return "", nil
	}

	if time.Now().After(item.expiry) {
		log.Printf("Key expired: %s", key)
		delete(s.data, key)
		return "", nil
	}

	hasher := sha1.New()
	hasher.Write([]byte(item.value))
	hash := hex.EncodeToString(hasher.Sum(nil))
	if hash != item.hash {
		log.Printf("Data integrity check failed for key: %s", key)
		return "", errors.New("data integrity check failed")
	}

	decryptedValue, err := util.Decrypt(item.value, s.key)
	if err != nil {
		log.Printf("Error decrypting value: %v", err)
		return "", err
	}

	log.Printf("Decrypted value for key %s: %s", key, decryptedValue)
	return string(decryptedValue), nil
}

func (s *Storage) cleanupExpired() {
	s.mu.Lock()
	defer s.mu.Unlock()

	for key, item := range s.data {
		if time.Now().After(item.expiry) {
			log.Printf("Cleaning up expired item with key: %s", key)
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

// GetAll returns all key-value pairs from the storage, decrypting the values before returning
func (s *Storage) GetAll() map[string]string {
	s.mu.Lock()
	defer s.mu.Unlock()

	allData := make(map[string]string)
	for key, item := range s.data {
		decryptedValue, err := util.Decrypt(item.value, s.key)
		if err != nil {
			log.Printf("Error decrypting value for key %s: %v", key, err)
			continue
		}
		allData[key] = string(decryptedValue)
	}
	return allData
}