package security

import (
	"crypto/rand"
	"fmt"
)

func Generate4DigitToken() (string, error) {
	b := make([]byte, 2)

	_, err := rand.Read(b)
	if err != nil {
		return "", fmt.Errorf("failed to generate token: %w", err)
	}

	n := int(b[0])<<8 + int(b[1])
	return fmt.Sprintf("%04d", n%10000), nil
}
