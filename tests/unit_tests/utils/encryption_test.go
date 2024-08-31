package tests

import (
	"testing"

	"gitlab.lrz.de/netintum/teaching/p2psec_projects_2024/DHT-14/pkg/util"
)

// TestGenerateRandomKey tests the GenerateRandomKey function
func TestGenerateRandomKey(t *testing.T) {
	key, err := util.GenerateRandomKey()
	if err != nil {
		t.Fatalf("Failed to generate random key: %v", err)
	}
	if len(key) != 32 {
		t.Errorf("Expected key length of 32, got %d", len(key))
	}
}

// TestEncryptDecrypt tests the Encrypt and Decrypt functions
func TestEncryptDecrypt(t *testing.T) {
	key, err := util.GenerateRandomKey()
	if err != nil {
		t.Fatalf("Failed to generate random key: %v", err)
	}

	originalData := []byte("This is a secret message")

	encryptedData, err := util.Encrypt(originalData, key)
	if err != nil {
		t.Fatalf("Failed to encrypt data: %v", err)
	}

	decryptedData, err := util.Decrypt(encryptedData, key)
	if err != nil {
		t.Fatalf("Failed to decrypt data: %v", err)
	}

	if string(decryptedData) != string(originalData) {
		t.Errorf("Decrypted data does not match original data: expected %s, got %s", string(originalData), string(decryptedData))
	}
}

// TestDecryptWithWrongKey tests the Decrypt function with a wrong key
func TestDecryptWithWrongKey(t *testing.T) {
	key, err := util.GenerateRandomKey()
	if err != nil {
		t.Fatalf("Failed to generate random key: %v", err)
	}

	wrongKey, err := util.GenerateRandomKey()
	if err != nil {
		t.Fatalf("Failed to generate wrong random key: %v", err)
	}

	originalData := []byte("This is a secret message")

	encryptedData, err := util.Encrypt(originalData, key)
	if err != nil {
		t.Fatalf("Failed to encrypt data: %v", err)
	}

	_, err = util.Decrypt(encryptedData, wrongKey)
	if err == nil {
		t.Fatalf("Decryption should have failed with wrong key")
	}
}