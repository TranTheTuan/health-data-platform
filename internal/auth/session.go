package auth

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"strings"
)

var ErrInvalidSignature = errors.New("invalid session signature")

// Sign creates a signed string in the format "value.signature"
func Sign(value string, secret string) string {
	h := hmac.New(sha256.New, []byte(secret))
	h.Write([]byte(value))
	signature := base64.RawURLEncoding.EncodeToString(h.Sum(nil))
	return value + "." + signature
}

// Verify checks the signature of a signed string and returns the original value if valid
func Verify(signedValue string, secret string) (string, error) {
	lastDot := strings.LastIndex(signedValue, ".")
	if lastDot == -1 {
		return "", ErrInvalidSignature
	}

	value := signedValue[:lastDot]
	signature := signedValue[lastDot+1:]

	h := hmac.New(sha256.New, []byte(secret))
	h.Write([]byte(value))
	expectedSignature := base64.RawURLEncoding.EncodeToString(h.Sum(nil))

	if !hmac.Equal([]byte(signature), []byte(expectedSignature)) {
		return "", ErrInvalidSignature
	}

	return value, nil
}
