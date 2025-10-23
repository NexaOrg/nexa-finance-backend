package utils

import (
	"os"
	"time"

	"github.com/golang-jwt/jwt/v4"
)

// GenerateJWT cria um token JWT com o ID do usuário e expiração de 24 horas
func GenerateJWT(userID string) (string, error) {
	// Define as claims (dados dentro do token)
	claims := jwt.MapClaims{
		"userId": userID,
		"exp":    time.Now().Add(24 * time.Hour).Unix(),
	}

	// Cria o token com as claims e método de assinatura
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	// Usa a chave secreta definida no .env
	secret := os.Getenv("JWT_SECRET")
	if secret == "" {
		secret = "default_secret" // fallback seguro para dev
	}

	// Assina e retorna o token
	return token.SignedString([]byte(secret))
}
