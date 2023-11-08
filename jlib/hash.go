package jlib

import (
	"crypto/sha256"
	"encoding/hex"
)

// Hash a input string into a sha256 string
// for deduplication purposes
func Hash(input string) string {
	hasher := sha256.New()

	hasher.Write([]byte(input))

	hashedBytes := hasher.Sum(nil)

	hashedString := hex.EncodeToString(hashedBytes)

	return hashedString
}
