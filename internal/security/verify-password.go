package security

import (
	"fmt"

	"golang.org/x/crypto/bcrypt"
)

const (
	Time       = 5
	Memory     = 64 * 1024
	Threads    = 4
	KeyLength  = 32
	SaltLength = 16
)

func VerifyPasswordMatch(plainPassword, storedHash string) error {
	err := bcrypt.CompareHashAndPassword([]byte(storedHash), []byte(plainPassword))
	if err != nil {
		return fmt.Errorf("senha inválida")
	}
	return nil
}
