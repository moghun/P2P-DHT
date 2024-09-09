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
	data         map[string]*storageItem
	mu           sync.Mutex
	cleanup_interval  time.Duration
	key          []byte // Encryption key
	cleanupTicker *time.Ticker // Holds the reference to the ticker
	stopCleanup   chan bool // Channel to signal stopping the cleanup routine
}

type storageItem struct {
	value  string
	expiry time.Time
	hash   string
}

func NewStorage(cleanup_interval time.Duration, key []byte) *Storage {
	storage := &Storage{
		data:        make(map[string]*storageItem),
		cleanup_interval:         cleanup_interval,
		key:         key,
		stopCleanup: make(chan bool),
	}

	storage.StartCleanup(cleanup_interval * time.Second)
	return storage
}

func (s *Storage) Put(key, value string, ttl int) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	util.Log().Printf("TTL: %d for the key (%s)", ttl, key)

	if (ttl < int(s.cleanup_interval)){
		ttl = int(s.cleanup_interval)
	}
	encryptedValue, err := util.Encrypt([]byte(value), s.key)
	if err != nil {
		util.Log().Errorf("Error encrypting value: %v", err)
		return err
	}
	util.Log().Infof("Encrypted value: %s", encryptedValue)

	hasher := sha1.New()
	hasher.Write([]byte(encryptedValue))
	hash := hex.EncodeToString(hasher.Sum(nil))
	util.Log().Infof("SHA1 hash of encrypted value: %s", hash)

	expiry := time.Now().Add(time.Duration(ttl) * time.Second)
	s.data[key] = &storageItem{
		value:  encryptedValue,
		expiry: expiry,
		hash:   hash,
	}
	util.Log().Infof("Stored item: key=%s, value=%s, expiry=%s, hash=%s", key, encryptedValue, expiry.String(), hash)
	return nil
}

func (s *Storage) Get(key string) (string, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	util.Log().Infof("Attempting to retrieve key: %x", key)

	item, exists := s.data[key]
	if !exists {
		util.Log().Infof("Key not found: %x", key)
		return "", nil
	}

	if time.Now().After(item.expiry) {
		util.Log().Infof("Key expired: %x", key)
		delete(s.data, key)
		return "", nil
	}

	hasher := sha1.New()
	hasher.Write([]byte(item.value))
	hash := hex.EncodeToString(hasher.Sum(nil))
	if hash != item.hash {
		util.Log().Infof("Data integrity check failed for key: %x", key)
		return "", errors.New("data integrity check failed")
	}

	decryptedValue, err := util.Decrypt(item.value, s.key)
	if err != nil {
		util.Log().Errorf("Error decrypting value: %v", err)
		return "", err
	}

	util.Log().Infof("Decrypted value for key %x: %s", key, decryptedValue)
	return string(decryptedValue), nil
}
func (s *Storage) CleanupExpired() {
	s.mu.Lock()
	defer s.mu.Unlock()

	for key, item := range s.data {
		if time.Now().After(item.expiry) {
			util.Log().Infof("Cleaning up expired item with key: %s", key)
			delete(s.data, key)
		}
	}
}

func (s *Storage) StartCleanup(interval time.Duration) {
	s.cleanupTicker = time.NewTicker(interval)
	go func() {
		for {
			select {
			case <-s.cleanupTicker.C:
				s.CleanupExpired()
			case <-s.stopCleanup:
				s.cleanupTicker.Stop() // Stop the ticker
				return
			}
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
			util.Log().Errorf("Error decrypting value for key %s: %v", key, err)
			continue
		}
		allData[key] = string(decryptedValue)
	}
	return allData
}

// StopCleanup stops the cleanup ticker and terminates the cleanup routine.
func (s *Storage) StopCleanup() {
	s.stopCleanup <- true
	log.Println("Storage cleanup routine stopped.")
}


/* FOR TEST PURPOSES*/
// Getter for the Data field
func (s *Storage) GetData() map[string]*storageItem {
	return s.data
}

// Getter for the Data field
func (s *storageItem) SetValue(setValue string){
	s.value = setValue
}

// Getter for the TTL field
func (s *Storage) GetTTL() time.Duration {
	return s.cleanup_interval
}

// Getter for the Key field
func (s *Storage) GetKey() []byte {
	return s.key
}

// Setter for the Key field (useful for testing purposes)
func (s *Storage) SetKey(key []byte) {
	s.key = key
}
/* FOR TEST PURPOSES*/