package tests

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"gitlab.lrz.de/netintum/teaching/p2psec_projects_2024/DHT-14/pkg/storage"
)

func TestNewStorage(t *testing.T) {
	key := []byte("testkey123456789")
	ttl := 24 * time.Hour

	store := storage.NewStorage(ttl, key)
	assert.NotNil(t, store)
	assert.Equal(t, key, store.GetKey())
	assert.Equal(t, ttl, store.GetTTL())
	assert.NotNil(t, store.GetData())
}

func TestStorage_PutAndGet(t *testing.T) {
	key := []byte("testkey123456789")
	store := storage.NewStorage(24 * time.Hour, key)

	err := store.Put("testkey", "testvalue", 3600)
	assert.NoError(t, err)

	value, err := store.Get("testkey")
	assert.NoError(t, err)
	assert.Equal(t, "testvalue", value)
}

func TestStorage_GetExpired(t *testing.T) {
	key := []byte("testkey123456789")
	store := storage.NewStorage(24 * time.Hour, key)

	err := store.Put("testkey", "testvalue", 1) // 1 second TTL
	assert.NoError(t, err)

	time.Sleep(2 * time.Second) // Wait for the item to expire

	value, err := store.Get("testkey")
	assert.NoError(t, err)
	assert.Empty(t, value)
}

func TestStorage_GetNonExistentKey(t *testing.T) {
	key := []byte("testkey123456789")
	store := storage.NewStorage(24 * time.Hour, key)

	value, err := store.Get("nonexistent")
	assert.NoError(t, err)
	assert.Empty(t, value)
}

func TestStorage_DataIntegrityCheck(t *testing.T) {
	key := []byte("testkey123456789")
	store := storage.NewStorage(24 * time.Hour, key)

	// Put a value
	err := store.Put("testkey", "testvalue", 3600)
	assert.NoError(t, err)

	// Manually tamper with the stored value
	store.GetData()["testkey"].SetValue("tamperedvalue")

	// Get the value and expect a data integrity error
	value, err := store.Get("testkey")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "data integrity check failed")
	assert.Empty(t, value)
}

func TestStorage_StartCleanup(t *testing.T) {
	key := []byte("testkey123456789")
	store := storage.NewStorage(1*time.Second, key)

	err := store.Put("testkey1", "testvalue1", 1) // 1 second TTL
	assert.NoError(t, err)
	err = store.Put("testkey2", "testvalue2", 3600) // 1 hour TTL
	assert.NoError(t, err)

	time.Sleep(2 * time.Second) // Wait for the cleanup to run

	// testkey1 should be cleaned up
	value, err := store.Get("testkey1")
	assert.NoError(t, err)
	assert.Empty(t, value)

	// testkey2 should still exist
	value, err = store.Get("testkey2")
	assert.NoError(t, err)
	assert.Equal(t, "testvalue2", value)
}

func TestStorage_GetAll(t *testing.T) {
	key := []byte("testkey123456789")
	store := storage.NewStorage(24 * time.Hour, key)

	err := store.Put("testkey1", "testvalue1", 3600)
	assert.NoError(t, err)
	err = store.Put("testkey2", "testvalue2", 3600)
	assert.NoError(t, err)

	allData := store.GetAll()
	assert.Len(t, allData, 2)
	assert.Equal(t, "testvalue1", allData["testkey1"])
	assert.Equal(t, "testvalue2", allData["testkey2"])
}

func TestStorage_GetKeyNotFound(t *testing.T) {
	key := []byte("testkey123456789")
	store := storage.NewStorage(24 * time.Hour, key)

	_, err := store.Get("nonexistent_key")
	assert.NoError(t, err)
}

func TestStorage_PutErrorOnEncrypt(t *testing.T) {
	// Invalid key length should cause encryption to fail
	key := []byte("shortkey")
	store := storage.NewStorage(24 * time.Hour, key)

	err := store.Put("testkey", "testvalue", 3600)
	assert.Error(t, err)
}

func TestStorage_GetErrorOnDecrypt(t *testing.T) {
	key := []byte("testkey123456789")
	store := storage.NewStorage(24 * time.Hour, key)

	err := store.Put("testkey", "testvalue", 3600)
	assert.NoError(t, err)

	// Tamper with the encryption key
	store.SetKey([]byte("wrongkey12345678"))

	_, err = store.Get("testkey")
	assert.Error(t, err)
}

func TestStorage_CleanupExpired(t *testing.T) {
	key := []byte("testkey123456789")
	store := storage.NewStorage(24 * time.Hour, key)

	// Add an item that expires immediately
	store.Put("expiringKey", "expiringValue", 1)
	time.Sleep(2 * time.Second)

	store.CleanupExpired()

	_, exists := store.GetData()["expiringKey"]
	assert.False(t, exists, "Expired item should have been cleaned up")
}
