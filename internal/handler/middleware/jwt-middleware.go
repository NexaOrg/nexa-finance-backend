package middleware

import (
	"fmt"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
)

func JWTMiddleware(c *fiber.Ctx) error {
	tokenString := c.Get("Authorization")
	idUser := c.Params("idUser")
	secret := "secret-key"

	token, response, err := parseToken(tokenString, secret)
	if err != nil {
		if err = c.Status(fiber.StatusUnauthorized).JSON(response); err != nil {
			return fmt.Errorf("filed to encode response: %s", err)
		}

		return nil
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		if err := c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"status":  401,
			"message": "Token inválido ou ausente.",
		}); err != nil {
			return fmt.Errorf("failed to send unauthorized response: %w", err)
		}

		return nil
	}

	tokenUserID, ok := claims["sub"].(string)
	if ok && tokenUserID != idUser {
		if err := c.Status(fiber.StatusForbidden).JSON(fiber.Map{
			"status":         403,
			"error":          "forbidden",
			"message":        "Acesso negado. idUser inconsistente.",
			"expectedIdUser": claims["sub"],
			"timestamp":      time.Now(),
		}); err != nil {
			return fmt.Errorf("ocorreu um erro interno: %w", err)
		}

		return nil
	}

	if err := c.Next(); err != nil {
		return fmt.Errorf("falha ao continar a requisição: %w", err)
	}

	return nil
}

func parseToken(tokenString string, secret string) (*jwt.Token, fiber.Map, error) {
	if tokenString == "" {
		response := fiber.Map{
			"status":    401,
			"error":     "unauthorized",
			"message":   "Token não fornecido ou inválido.",
			"timestamp": time.Now(),
		}

		return nil, response, fmt.Errorf("empty token")
	}

	token, err := jwt.Parse(tokenString, func(_ *jwt.Token) (interface{}, error) {
		return []byte(secret), nil
	})

	if err != nil || !token.Valid {
		response := fiber.Map{
			"status":    401,
			"error":     "unauthorized",
			"message":   "Acesso negado. Token inválido.",
			"timestamp": time.Now(),
		}

		return nil, response, fmt.Errorf("invalid token: %w", err)
	}

	return token, nil, nil
}
