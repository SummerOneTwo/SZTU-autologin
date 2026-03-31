package main

import (
	"crypto/md5"
	"crypto/sha1"
	"encoding/base64"
	"encoding/hex"
)

// getMD5 calculates MD5 hash of string with optional token suffix
func getMD5(s string, token string) string {
	h := md5.New()
	h.Write([]byte(s))
	if token != "" {
		h.Write([]byte(token))
	}
	return hex.EncodeToString(h.Sum(nil))
}

// getSHA1 calculates SHA1 hash of string
func getSHA1(s string) string {
	h := sha1.New()
	h.Write([]byte(s))
	return hex.EncodeToString(h.Sum(nil))
}

// getBase64 encodes bytes to base64 string
func getBase64(data []byte) string {
	return base64.StdEncoding.EncodeToString(data)
}

// getXEncode implements SRUN portal XEncode algorithm
// TODO: This is a placeholder. Full implementation in Task 9.
// For now, returns the input string as bytes (will break login)
func getXEncode(s string, key string) []byte {
	// Placeholder implementation
	// The real SRUN XEncode algorithm is complex and will be implemented in Task 9
	return []byte(s)
}
