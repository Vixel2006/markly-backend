package utils

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"io"
)

// GenerateCodeVerifier generates a cryptographically random string for PKCE.
func GenerateCodeVerifier() (string, error) {
	b := make([]byte, 32) // 32 bytes for a 43-character base64 string
	if _, err := io.ReadFull(rand.Reader, b); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(b), nil
}

// GenerateCodeChallenge generates the S256 code challenge from a code verifier.
func GenerateCodeChallenge(codeVerifier string) string {
	h := sha256.New()
	h.Write([]byte(codeVerifier))
	return base64.RawURLEncoding.EncodeToString(h.Sum(nil))
}
