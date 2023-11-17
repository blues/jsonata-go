package jlib

import (
	"crypto/md5"
	"crypto/sha256"
	"encoding/hex"
)

// Hash a input string into a md5 string
// for deduplication purposes
func HashMD5(input string) string {
	hash := md5.Sum([]byte(input))

	hashedString := hex.EncodeToString(hash[:])

	return hashedString
}

// Hash a input string into a sha256 string
// for deduplication purposes
func Hash256(input string) string {
	hasher := sha256.New()

	hasher.Write([]byte(input))

	hashedBytes := hasher.Sum(nil)

	hashedString := hex.EncodeToString(hashedBytes)

	return hashedString
}
