package security

import (
	"fmt"

	"golang.org/x/crypto/bcrypt"
)

func VerifyPasswordMatch(plainPassword, storedHash string) error {
	err := bcrypt.CompareHashAndPassword([]byte(storedHash), []byte(plainPassword))
	if err != nil {
		return fmt.Errorf("senha inválida")
	}
	return nil
}
