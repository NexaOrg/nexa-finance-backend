package security

import "golang.org/x/crypto/bcrypt"

func EncryptPassword(password string) ([]byte, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.MinCost)
	if err != nil {
		return nil, err
	}

	return hash, nil
}
