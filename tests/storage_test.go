package tests

import (
	"log"
	"testing"
	"time"

	"gitlab.lrz.de/netintum/teaching/p2psec_projects_2024/DHT-14/pkg/storage"
	"gitlab.lrz.de/netintum/teaching/p2psec_projects_2024/DHT-14/pkg/util"
)

func TestStorage(t *testing.T) {
	key, _ := util.GenerateRandomKey()
	storageInstance := storage.NewStorage(10*time.Second, key)

	err := storageInstance.Put("key1", "value1", 5)
	if err != nil {
		t.Fatalf("Failed to put value: %v", err)
	}

	val, err := storageInstance.Get("key1")
	if err != nil {
		t.Fatalf("Failed to get value: %v", err)
	}
	if val != "value1" {
		t.Errorf("Expected value1, got %s", val)
	}

	log.Println("Sleeping for 6 seconds to allow item to expire...")
	time.Sleep(6 * time.Second)
	val, err = storageInstance.Get("key1")
	if err == nil && val != "" {
		t.Errorf("Expected value to be expired, but got %s", val)
	}
}