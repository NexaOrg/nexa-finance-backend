package handler

import (
	"bytes"
	"fmt"
	"nexa/internal/factory"
	"nexa/internal/model"
	"nexa/internal/repository"
	"nexa/internal/security"
	"nexa/internal/utils"
	"os"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt"
	"github.com/jackc/pgx/v5"
)

type UserAuthenticationHandler struct {
	UserRepository                 *repository.UserRepository
	UserAuthenticationTokenRepo    *repository.UserAuthenticationTokenRepository
	UserAuthenticationTokenBuilder *factory.UserAuthenticationTokenFactory
	MailServer                     *utils.MailServer
}

func NewUserAuthenticationHandler(db *pgx.Conn, mailServer *utils.MailServer) *UserAuthenticationHandler {
	return &UserAuthenticationHandler{
		UserRepository:                 repository.NewUserRepository(db),
		UserAuthenticationTokenRepo:    repository.NewUserAuthenticationTokenRepository(db, "db_nexa", "tb_user_authentication_token"),
		UserAuthenticationTokenBuilder: factory.NewUserAuthenticationTokenFactory(),
		MailServer:                     mailServer,
	}
}

func (ua *UserAuthenticationHandler) HandleInitialAuthentication(userID string) error {
	token, err := ua.createUserAuthenticationToken(userID)
	if err != nil {
		return err
	}
	_, err = ua.sendAuthenticationEmail(token.Code, userID)
	return err
}

func (ua *UserAuthenticationHandler) VerifyUser(c *fiber.Ctx) error {
	userID, code, err := ua.getUserIDAndCode(c)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}

	token, err := ua.UserAuthenticationTokenRepo.FindTokenByUserID(userID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	codeStatus, err := ua.validateTokenAndCreateNewIfNeeded(token, userID, code)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": codeStatus})
	}

	if err := ua.activateUserAccount(userID); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	// Gere o JWT com o ID do usuário como sub
	tokenStr, err := utils.GenerateJWT(userID)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Erro ao gerar token"})
	}

	return c.JSON(fiber.Map{"token": tokenStr})
}

func (ua *UserAuthenticationHandler) activateUserAccount(userID string) error {
	return ua.UserRepository.UpdateByID(userID, map[string]interface{}{"is_active": true})
}

func (ua *UserAuthenticationHandler) getUserIDAndCode(c *fiber.Ctx) (string, string, error) {
	var request struct {
		UserID string `json:"idUser"`
		Code   string `json:"code"`
	}
	if err := c.BodyParser(&request); err != nil {
		return "", "", err
	}
	if request.UserID == "" || request.Code == "" {
		return "", "", fmt.Errorf("userID or code missing")
	}
	return request.UserID, request.Code, nil
}

func (ua *UserAuthenticationHandler) validateTokenAndCreateNewIfNeeded(token *model.UserAuthenticationToken, userID, code string) (string, error) {
	if token == nil || token.HasExpired() || token.Fails > 2 {
		var tokenID string
		if token != nil {
			tokenID = token.ID
		}
		HTTPError, err := ua.deleteAndCreateNewUserAuthenticationToken(tokenID, userID)
		if err != nil {
			return HTTPError, err
		}
		return "EXPIRED_AUTHENTICATION_TOKEN", fmt.Errorf("token expirado")
	}
	if token.Code != code {
		if err := ua.UserAuthenticationTokenRepo.IncrementFails(token.ID); err != nil {
			return "INTERNAL_SERVER_ERROR", err
		}
		return "INVALID_USER_AUTHENTICATION_TOKEN", fmt.Errorf("código inválido")
	}
	return "", nil
}

func (ua *UserAuthenticationHandler) deleteAndCreateNewUserAuthenticationToken(tokenID, userID string) (string, error) {
	if tokenID != "" {
		_ = ua.UserAuthenticationTokenRepo.Delete(tokenID)
	}
	token, err := ua.createUserAuthenticationToken(userID)
	if err != nil {
		return "INTERNAL_SERVER_ERROR", err
	}
	HTTPError, err := ua.sendAuthenticationEmail(token.Code, userID)
	if err != nil {
		return HTTPError, err
	}
	return "", nil
}

func (ua *UserAuthenticationHandler) createUserAuthenticationToken(userID string) (*model.UserAuthenticationToken, error) {
	code, err := utils.GenerateCode(4)
	if err != nil {
		return nil, err
	}
	token := ua.UserAuthenticationTokenBuilder.CreateUserAuthenticationToken(userID, code, 1440)
	// Insere e obtém ID gerado
	id, err := ua.UserAuthenticationTokenRepo.Insert(token)
	if err != nil {
		return nil, err
	}
	token.ID = id
	return token, nil
}

func (ua *UserAuthenticationHandler) sendAuthenticationEmail(code string, userID string) (string, error) {
	user, err := ua.UserRepository.FindByFilter("id", userID)
	if err != nil {
		return "INTERNAL_SERVER_ERROR", err
	}
	if user == nil {
		return "NOT_FOUND", fmt.Errorf("user not found")
	}

	var body bytes.Buffer
	template, err := utils.ParseFile("assets/index.html")
	if err != nil {
		return "INTERNAL_SERVER_ERROR", err
	}

	if err = template.Execute(&body, struct {
		Name string
		Code string
	}{Name: user.Name, Code: code}); err != nil {
		return "INTERNAL_SERVER_ERROR", err
	}

	if err = ua.MailServer.SendEmailHTML(
		"Validação de E-mail",
		body.String(),
		[]string{user.Email},
	); err != nil {
		return "INTERNAL_SERVER_ERROR", err
	}

	return "", nil
}

func (ua *UserAuthenticationHandler) GetPublicKey(c *fiber.Ctx) error {
	pubPEM, err := security.LoadPublicKeyPEMFlatString()
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Erro ao carregar chave pública"})
	}
	return c.JSON(fiber.Map{"publicKey": pubPEM})
}

var secretKey = []byte("secret-key")

func (ua *UserAuthenticationHandler) CreateToken(id string, issuer string) (string, error) {
	secretKey := []byte(os.Getenv("JWT_SECRET"))

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"sub": id,
		"iss": issuer,
		"iat": time.Now().Unix(),
		"exp": time.Now().Add(2 * time.Hour).Unix(),
	})

	tokenString, err := token.SignedString(secretKey)
	if err != nil {
		return "", fmt.Errorf("failed to stringify token: %w", err)
	}

	return tokenString, nil
}
