package auth

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
)

func GenerateSessionToken() (string, error) {
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return "", fmt.Errorf("failed to generate session token: %w", err)
	}
	return base64.URLEncoding.EncodeToString(bytes), nil
}
