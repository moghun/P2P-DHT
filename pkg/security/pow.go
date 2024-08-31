// pkg/security/pow.go

package security

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"strings"
)

// Difficulty level for PoW (number of leading zeros required in the hash)
const difficulty = 4

// GenerateNodeIDWithPoW generates a node ID with a proof-of-work mechanism.
func GenerateNodeIDWithPoW(ip string, port int) (string, int) {
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

// ValidateNodeIDWithPoW validates that the node ID meets the required PoW difficulty.
func ValidateNodeIDWithPoW(ip string, port int, id string, nonce int) bool {
	data := fmt.Sprintf("%s:%d:%d", ip, port, nonce)
	hash := sha256.Sum256([]byte(data))
	hashStr := hex.EncodeToString(hash[:])

	return strings.HasPrefix(hashStr, strings.Repeat("0", difficulty)) && hashStr == id
}
