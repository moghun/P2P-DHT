// pkg/security/pow.go

package security

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"strings"
)

// GenerateNodeIDWithPoW generates a node ID with a proof-of-work mechanism, now with dynamic difficulty.
func GenerateNodeIDWithPoW(ip string, port int, difficulty int) (string, int) {
	nonce := 0
	for {
		data := fmt.Sprintf("%s:%d:%d", ip, port, nonce)
		hash := sha256.Sum256([]byte(data))
		hashStr := hex.EncodeToString(hash[:])

		if strings.HasPrefix(hashStr, strings.Repeat("0", difficulty)) {
			return hashStr, nonce
		}

		nonce++
	}
}

// ValidateNodeIDWithPoW validates that the node ID meets the required PoW difficulty, now with dynamic difficulty.
func ValidateNodeIDWithPoW(ip string, port int, id string, nonce int, difficulty int) bool {
	data := fmt.Sprintf("%s:%d:%d", ip, port, nonce)
	hash := sha256.Sum256([]byte(data))
	hashStr := hex.EncodeToString(hash[:])

	return strings.HasPrefix(hashStr, strings.Repeat("0", difficulty)) && hashStr == id
}