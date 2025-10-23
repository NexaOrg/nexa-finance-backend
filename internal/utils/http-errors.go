package utils

import (
	"fmt"
	"net/http"
	"nexa/internal/model"
	"time"

	"github.com/gofiber/fiber/v2"
)

func EncodeRequestError(c *fiber.Ctx, errorType, path string) error {
	errorResponse, exists := HTTPErrors[errorType]
	if !exists {
		if err := c.SendStatus(404); err != nil {
			return fmt.Errorf("failed to send 404 status: %w", err)
		}

		return fmt.Errorf("error not found: %s", errorType)
	}

	errorResponse.Path = path

	if err := c.Status(errorResponse.StatusCode).JSON(errorResponse); err != nil {
		return fmt.Errorf("failed to send JSON with status %d: %w", errorResponse.StatusCode, err)
	}

	return nil
}

var HTTPErrors = map[string]model.ErrorResponse{
	"INVALID_ID_USER": {
		StatusCode: 400,
		Error:      "Bad Request",
		Message:    "formato de IDUser inválido",
		Timestamp:  time.Now(),
	},
	"INVALID_ID_QUIZ": {
		StatusCode: 400,
		Error:      "Bad Request",
		Message:    "formato de IDQuiz inválido",
		Timestamp:  time.Now(),
	},
	"USER_QUIZ_NOT_FOUND": {
		StatusCode: 404,
		Error:      "Not Found",
		Message:    "User or quiz not found.",
		Timestamp:  time.Now(),
	},
	"REQUIRED_EMAIL": {
		StatusCode: 400,
		Error:      "Bad Request",
		Message:    "email is required.",
		Timestamp:  time.Now(),
		Input:      "email",
	},
	"USER_ALREADY_IN_WAIT_LIST": {
		StatusCode: 409,
		Error:      "Conflict",
		Message:    "email already registered",
		Timestamp:  time.Now(),
		Input:      "email",
	},
	"USER_NOT_ALLOWED": {
		StatusCode: 401,
		Error:      "Unauthorized",
		Message:    "wait-list",
		Timestamp:  time.Now(),
	},
	"USER_ALEADY_REGISTERED": {
		StatusCode: 403,
		Error:      "Forbidden",
		Message:    "E-mail já está em uso",
		Timestamp:  time.Now(),
		Input:      "email",
	},
	"USER_NOT_ACTIVE": {
		StatusCode: 400,
		Error:      "Bad Request",
		Message:    "This user is not active.",
		Timestamp:  time.Now(),
		Path:       "/login",
	},
	"USER_ALREADY_ACTIVE": {
		StatusCode: fiber.StatusBadRequest,
		Error:      "Bad Request",
		Message:    "Usuário já está ativo.",
		Timestamp:  time.Now(),
	},
	"INVALID_PHOTO": {
		StatusCode: http.StatusBadRequest,
		Error:      "Bad Request",
		Message:    "failed to get image",
		Timestamp:  time.Now(),
		Input:      "photo",
	},
	"INVALID_LOGIN_CREDENTIALS": {
		StatusCode: http.StatusBadRequest,
		Error:      "Bad Request",
		Message:    "Credenciais não correspondentes",
		Timestamp:  time.Now(),
		Input:      "email/password",
	},
	"INVALID_USER_ID": {
		StatusCode: fiber.StatusBadRequest,
		Error:      "Bad Request",
		Message:    "Dados de entrada inválidos. Verifique o formato e tente novamente.",
		Timestamp:  time.Now(),
	},
	"INTERNAL_SERVER_ERROR": {
		StatusCode: fiber.StatusInternalServerError,
		Error:      "Internal Server Error",
		Message:    "Ocorreu um erro no servidor.",
		Timestamp:  time.Now(),
	},
	"USER_NOT_FOUND": {
		StatusCode: fiber.StatusNotFound,
		Error:      "Not Found",
		Message:    "Usuário não encontrado.",
		Timestamp:  time.Now(),
	},
	"INVALID_USER_AUTHENTICATION_TOKEN": {
		StatusCode: fiber.StatusBadRequest,
		Error:      "Bad Request",
		Message:    "Código inválido.",
		Timestamp:  time.Now(),
	},
	"EXPIRED_AUTHENTICATION_TOKEN": {
		StatusCode: fiber.StatusBadRequest,
		Error:      "Bad Request",
		Message:    "Token de autenticação expirado.",
		Timestamp:  time.Now(),
	},
	"EMPTY_USER": {
		StatusCode: 400,
		Error:      "Bad Request",
		Message:    "empty user object is not allowed",
		Timestamp:  time.Now(),
		Input:      "user",
	},
	"FUTURE_DATE": {
		StatusCode: 400,
		Error:      "Bad Request",
		Message:    "a future date is not allowed",
		Timestamp:  time.Now(),
		Input:      "date",
	},
	"MORE_THAN_500_CHARS": {
		StatusCode: 400,
		Error:      "Bad Request",
		Message:    "O nome deve conter no máximo 500 caracteres",
		Timestamp:  time.Now(),
		Input:      "name",
	},
	"INVALID_AGE": {
		StatusCode: 400,
		Error:      "Bad Request",
		Message:    "Idade mínima de 18 anos.",
		Timestamp:  time.Now(),
		Input:      "date",
	},
	"INVALID_DATE_FORMAT": {
		StatusCode: 400,
		Error:      "Bad Request",
		Message:    "Formato de data inválido.",
		Timestamp:  time.Now(),
		Input:      "date",
	},
	"INVALID_NAME": {
		StatusCode: 400,
		Error:      "Bad Request",
		Message:    "Nome não aceita números ou caracteres especiais.",
		Timestamp:  time.Now(),
		Input:      "name",
	},
	"REQUIRED_NAME": {
		StatusCode: 400,
		Error:      "Bad Request",
		Message:    "name is required",
		Timestamp:  time.Now(),
		Input:      "name",
	},
	"INVALID_PASSWORD": {
		StatusCode: 400,
		Error:      "Bad Request",
		Message:    "A senha deve ter no mínimo 8 caracteres, sendo uma letra maiúscula, uma letra minúscula, um caractere especial e um número",
		Timestamp:  time.Now(),
		Input:      "password",
	},
	"EMAIL_ALEADY_REGISTERED": {
		StatusCode: http.StatusBadRequest,
		Error:      "Bad Request",
		Message:    "E-mail já está em uso",
		Timestamp:  time.Now(),
		Input:      "email",
	},
	"INVALID_EMAIL": {
		StatusCode: http.StatusBadRequest,
		Error:      "Bad Request",
		Message:    "E-mail inválido",
		Timestamp:  time.Now(),
		Input:      "email",
	},
	"INVALID_BODY_FORMAT": {
		StatusCode: http.StatusBadRequest,
		Error:      "Bad Request",
		Message:    "Formato de JSON inválido",
		Timestamp:  time.Now(),
	},
}
