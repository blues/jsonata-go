package jlib

import (
	"crypto/md5"
	"crypto/sha256"
	"encoding/hex"
	"strings"
)

// Hash a input string into a sha256 string
// for deduplication purposes
func Hash(hashType, input string) string {
	switch strings.ToLower(hashType) {
	case "sha256":
		hasher := sha256.New()

		hasher.Write([]byte(input))

		hashedBytes := hasher.Sum(nil)

		hashedString := hex.EncodeToString(hashedBytes)

		return hashedString
	case "md5":
		hash := md5.Sum([]byte(input))

		hashedString := hex.EncodeToString(hash[:])

		return hashedString
	}

	hash := md5.Sum([]byte(input))

	hashedString := hex.EncodeToString(hash[:])

	return hashedString
}
