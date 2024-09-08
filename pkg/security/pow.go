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

// GenerateNodeIDWithPoW generates a 160-bit node ID with a proof-of-work mechanism.
func GenerateNodeIDWithPoW(ip string, port int) (string, int) {
	nonce := 0
	for {
		// Combine IP, port, and nonce to create unique input
		data := fmt.Sprintf("%s:%d:%d", ip, port, nonce)
		hash := sha256.Sum256([]byte(data))

		// Truncate to the first 160 bits (20 bytes) of the SHA-256 hash
		truncatedHash := hash[:20]

		// Convert truncated hash to hexadecimal string
		hashStr := hex.EncodeToString(truncatedHash)

		// Check if hash satisfies the difficulty condition
		if strings.HasPrefix(hashStr, strings.Repeat("0", difficulty)) {
			return hashStr, nonce
		}

		// Increment nonce for next iteration
		nonce++
	}
}

// ValidateNodeIDWithPoW validates that the node ID meets the required PoW difficulty.
func ValidateNodeIDWithPoW(ip string, port int, id string, nonce int) bool {
	data := fmt.Sprintf("%s:%d:%d", ip, port, nonce)
	hash := sha256.Sum256([]byte(data))

	// Truncate to the first 160 bits (20 bytes) of the SHA-256 hash
	truncatedHash := hash[:20]

	// Convert truncated hash to hexadecimal string
	hashStr := hex.EncodeToString(truncatedHash)

	// Validate both the difficulty condition and that the generated hash matches the given ID
	return strings.HasPrefix(hashStr, strings.Repeat("0", difficulty)) && hashStr == id
}
