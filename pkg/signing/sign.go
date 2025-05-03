// Package signing provides functions for creating and validating secure URL signatures
// using HMAC-SHA256.
package signing

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"fmt"
)

// Sign computes a URL-safe, truncated HMAC-SHA256 signature.
//
//	keyHex:  	hex-encoded HMAC key
//	saltHex: 	hex-encoded salt
//	content:  the content to sign
//	size:    	number of bytes to keep from the HMAC digest (max 32)
//
// Returns the URL-safe Base64 signature, or an error if hex decoding fails
func Sign(keyHex string, saltHex string, content string, size int) (string, error) {
	// Decode the hex-encoded key
	key, err := hex.DecodeString(keyHex)
	if err != nil {
		return "", fmt.Errorf("invalid key hex: %w", err)
	}

	// Decode the hex-encoded salt
	salt, err := hex.DecodeString(saltHex)
	if err != nil {
		return "", fmt.Errorf("invalid salt hex: %w", err)
	}

	// Create HMAC-SHA256 with the decoded key
	mac := hmac.New(sha256.New, key)

	// Feed in the salt and then the target
	mac.Write(salt)
	mac.Write([]byte(content))

	// Compute full SHA-256 HMAC
	fullDig := mac.Sum(nil)

	// Truncate to `size` bytes
	if size < 0 || size > len(fullDig) {
		size = len(fullDig)
	}
	truncDig := fullDig[:size]

	// URL-safe Base64 encode without padding
	sig := base64.RawURLEncoding.EncodeToString(truncDig)
	return sig, nil
}

// UrlSafeEncode encodes data using URL-safe Base64 encoding without padding.
//
// This is useful for encoding binary data in a URL-friendly format.
//
// Parameters:
//   - data: The byte slice to encode.
//
// Returns the URL-safe Base64 encoded string.
func UrlSafeEncode(data []byte) string {
	return base64.RawURLEncoding.EncodeToString(data)
}

// UrlSafeDecode decodes a URL-safe Base64 encoded string.
//
// Parameters:
//   - encodedStr: The URL-safe Base64 encoded string to decode.
//
// Returns the decoded bytes and an error if decoding fails.
func UrlSafeDecode(encodedStr string) ([]byte, error) {
	return base64.RawURLEncoding.DecodeString(encodedStr)
}
