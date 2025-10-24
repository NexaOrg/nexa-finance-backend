package handler

import (
	"bytes"
	"fmt"
	"nexa/internal/factory"
	"nexa/internal/model"
	"nexa/internal/repository"
	"nexa/internal/security"
	"nexa/internal/utils"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt"
	"github.com/jackc/pgx/v5"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type UserAuthenticationHandler struct {
	UserRepository                 *repository.UserRepository
	UserAuthenticationTokenRepo    *repository.UserAuthenticationTokenRepository
	UserAuthenticationTokenBuilder *factory.UserAuthenticationTokenFactory
	MailServer                     *utils.MailServer
}

func NewUserAuthenticationHandler(db *pgx.Conn, mailServer *utils.MailServer) *UserAuthenticationHandler {
	return &UserAuthenticationHandler{
		UserRepository: repository.NewUserRepository(db),
		//UserAuthenticationTokenRepo:    repository.NewUserAuthenticationTokenRepository(client, "Cluster0", "user_authentication_tokens"),
		UserAuthenticationTokenBuilder: factory.NewUserAuthenticationTokenFactory(),
		MailServer:                     mailServer,
	}
}

func (ua *UserAuthenticationHandler) HandleInitialAuthentication(userID primitive.ObjectID) error {
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

	tokenStr, err := utils.GenerateJWT(userID.Hex())
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Erro ao gerar token"})
	}

	return c.JSON(fiber.Map{"token": tokenStr})
}

func (ua *UserAuthenticationHandler) activateUserAccount(userObjectID primitive.ObjectID) error {
	return ua.UserRepository.UpdateByID(userObjectID, map[string]interface{}{"isActive": true})
}

func (ua *UserAuthenticationHandler) getUserIDAndCode(c *fiber.Ctx) (primitive.ObjectID, string, error) {
	var request struct {
		UserID string `json:"idUser"`
		Code   string `json:"code"`
	}
	if err := c.BodyParser(&request); err != nil {
		return primitive.NilObjectID, "", err
	}
	userObjectID, err := primitive.ObjectIDFromHex(request.UserID)
	if err != nil {
		return primitive.NilObjectID, "", err
	}
	return userObjectID, request.Code, nil
}

func (ua *UserAuthenticationHandler) validateTokenAndCreateNewIfNeeded(token *model.UserAuthenticationToken, userID primitive.ObjectID, code string) (string, error) {
	if token == nil || token.HasExpired() || token.Fails > 2 {
		HTTPError, err := ua.deleteAndCreateNewUserAuthenticationToken(token.ID, userID)
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

func (ua *UserAuthenticationHandler) deleteAndCreateNewUserAuthenticationToken(tokenID, userID primitive.ObjectID) (string, error) {
	if err := ua.UserAuthenticationTokenRepo.Delete(tokenID); err != nil {
		return "INTERNAL_SERVER_ERROR", err
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

func (ua *UserAuthenticationHandler) createUserAuthenticationToken(userID primitive.ObjectID) (*model.UserAuthenticationToken, error) {
	code, err := utils.GenerateCode(4)
	if err != nil {
		return nil, err
	}
	token := ua.UserAuthenticationTokenBuilder.CreateUserAuthenticationToken(userID, code, 1440)
	id, err := ua.UserAuthenticationTokenRepo.Insert(token)
	if err != nil {
		return nil, err
	}
	token.ID = id
	return token, nil
}

func (ua *UserAuthenticationHandler) sendAuthenticationEmail(code string, userID primitive.ObjectID) (string, error) {
	user, err := ua.UserRepository.FindByFilter("_id", userID)
	if err != nil {
		return "INTERNAL_SERVER_ERROR", err
	}
	var body bytes.Buffer
	template, err := utils.ParseFile("assets/index.html")
	if err != nil {
		return "INTERNAL_SERVER_ERROR", err
	}
	if err = template.Execute(&body, struct {
		Name string
		Code string
	}{Name: user.FirstName, Code: code}); err != nil {
		return "INTERNAL_SERVER_ERROR", err
	}
	if err = ua.MailServer.SendEmailHTML("Validação de E-mail", body.String(), []string{user.Email}); err != nil {
		return "INTERNAL_SERVER_ERROR", err
	}
	return "", nil
}

func (ua *UserAuthenticationHandler) GetPublicKey(c *fiber.Ctx) error {
	// Agora retorna a chave PEM diretamente
	pubPEM, err := security.LoadPublicKeyPEMFlatString()
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Erro ao carregar chave pública"})
	}
	return c.JSON(fiber.Map{"publicKey": pubPEM})
}

var secretKey = []byte("secret-key")

func (ua *UserAuthenticationHandler) CreateToken(id primitive.ObjectID, issuer string) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256,
		jwt.MapClaims{
			"sub": id.Hex(),
			"iss": issuer,
			"iat": time.Now().Unix(),
			"exp": time.Now().Add(time.Hour * 2).Unix(),
		})

	tokenString, err := token.SignedString(secretKey)
	if err != nil {
		return "", fmt.Errorf("failed to stringfy token")
	}

	return tokenString, nil
}
