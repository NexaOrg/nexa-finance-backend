package utils

import (
	"crypto/rand"
	"encoding/base32"
	"fmt"
)

func GenerateCode(length int) (string, error) {
	randomBytes := make([]byte, 32)

	_, err := rand.Read(randomBytes)
	if err != nil {
		return "", fmt.Errorf("failed to generate code: %s", err)
	}

	return base32.StdEncoding.EncodeToString(randomBytes)[:length], nil
}
