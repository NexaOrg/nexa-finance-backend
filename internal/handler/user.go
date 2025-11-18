package handler

import (
	"bytes"
	"crypto/sha1"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/url"
	"nexa/internal/factory"
	"nexa/internal/model"
	"nexa/internal/repository"
	"nexa/internal/security"
	"nexa/internal/utils"
	"os"
	"regexp"
	"strings"
	"time"

	"github.com/dlclark/regexp2"
	"github.com/gofiber/fiber/v2"
	"github.com/jackc/pgx/v5"
	"github.com/rs/zerolog/log"
)

type UserHandler struct {
	UserFactory               *factory.UserFactory
	UserRepository            *repository.UserRepository
	UserAuthenticationHandler *UserAuthenticationHandler
}

func NewUserHandler(db *pgx.Conn) *UserHandler {
	return &UserHandler{
		UserRepository:            repository.NewUserRepository(db),
		UserFactory:               factory.NewUserFactory(),
		UserAuthenticationHandler: NewUserAuthenticationHandler(db, nil),
	}
}

func (u *UserHandler) RegisterUser(c *fiber.Ctx) error {
	var modelUser model.User

	if err := c.BodyParser(&modelUser); err != nil {
		return c.Status(400).JSON(utils.EncodeRequestError(c, "INVALID_BODY_FORMAT"))
	}

	if len(modelUser.Name) > 20 {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid name", "message": "O nome não deve conter mais de 20 caracteres"})
	}

	passwordRegex := regexp2.MustCompile(`^(?=.*[A-Z])(?=.*\d)(?=.*[!@#\$%\^&\*\(\)_\+\-=\[\]{};':"\\|,.<>\/?]).{6,}$`, 0)
	match, _ := passwordRegex.MatchString(modelUser.Password)

	if !match {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid password", "message": "A senha deve possuir no mínimo 6 caracteres, contendo uma letra maiúscula, um número e um caractere especial"})
	}

	emailRegex := regexp2.MustCompile(`^[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}$`, 0)
	match, _ = emailRegex.MatchString(modelUser.Email)

	if !match {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid email", "message": "O email inserido não é válido"})
	}

	modelUser.LastLogin = time.Now()

	hash, err := security.EncryptPassword(modelUser.Password)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"message": "Internal Server Error", "error": err.Error()})
	}

	modelUser.Password = string(hash)

	err = u.UserRepository.InsertUser(&modelUser)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"message": "Internal Server Error", "error": err.Error()})
	}

	return c.Status(201).JSON(fiber.Map{"message": "User creation successful"})
}

func (u *UserHandler) validateUser(user *model.User) (bool, error) {
	return false, nil
}

func (u *UserHandler) validateEmail(email string) (*string, string, error) {
	return nil, "", nil
}

func (u *UserHandler) ValidateName(name string) (string, error) {
	const MaxLength = 200
	if name == "" {
		return "REQUIRED_NAME", fmt.Errorf("nome é obrigatório")
	}

	regex := regexp.MustCompile(`^[a-zA-ZÀ-ÖØ-öø-ÿ\s]+$`)
	if !regex.MatchString(name) {
		return "INVALID_NAME", fmt.Errorf("nome contém caracteres inválidos")
	}

	formattedName := utils.FormatName(name)
	if len(formattedName) > MaxLength {
		return "MORE_THAN_500_CHARS", fmt.Errorf("nome muito longo")
	}

	return "", nil
}

func (u *UserHandler) LoginUser(c *fiber.Ctx) error {
	var user model.User

	_, err := utils.ParseBody(c, &user)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":   "INVALID_BODY_FORMAT",
			"message": fmt.Sprintf("failed to parse body: %s", err),
		})
	}

	email := strings.ToLower(strings.TrimSpace(user.Email))
	if email == "" || user.Password == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":   "INVALID_CREDENTIALS",
			"message": "Email e senha são obrigatórios",
		})
	}

	dbUser, err := u.UserRepository.FindByFilter("email", email)
	if err != nil {
		log.Error().Err(err).Msg("failed to query user")
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":   "INTERNAL_SERVER_ERROR",
			"message": "Erro ao buscar usuário no banco de dados",
		})
	}

	if dbUser == nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error":   "INVALID_CREDENTIALS",
			"message": "Email ou senha incorretos",
		})
	}

	if err := u.validateLoginCredentials(dbUser, user.Password); err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error":   "INVALID_CREDENTIALS",
			"message": "Email ou senha incorretos",
		})
	}

	if !dbUser.IsActive {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
			"status":  fiber.StatusForbidden,
			"message": "Este usuário está inativo.",
			"idUser":  dbUser.ID,
		})
	}

	existing, err := u.UserAuthenticationHandler.UserAuthenticationTokenRepo.FindTokenByUserID(dbUser.ID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":   "INTERNAL_SERVER_ERROR",
			"message": "Falha ao buscar token existente",
		})
	}

	if existing != nil {
		_ = u.UserAuthenticationHandler.UserAuthenticationTokenRepo.Delete(existing.ID)
	}

	code, err := utils.GenerateCode(6)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":   "INTERNAL_SERVER_ERROR",
			"message": "Falha ao gerar código de autenticação",
		})
	}

	tokenData := u.UserAuthenticationHandler.UserAuthenticationTokenBuilder.CreateUserAuthenticationToken(dbUser.ID, code, 1440)

	tokenID, err := u.UserAuthenticationHandler.UserAuthenticationTokenRepo.Insert(tokenData)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":   "INTERNAL_SERVER_ERROR",
			"message": "Falha ao armazenar token no banco",
		})
	}
	tokenData.ID = tokenID

	jwtToken, err := u.UserAuthenticationHandler.CreateToken(dbUser.ID, "/auth/login")
	if err != nil {
		log.Error().Err(err).Msg("failed to create JWT token")
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":   "INTERNAL_SERVER_ERROR",
			"message": "Falha ao gerar token de autenticação",
		})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"status":      fiber.StatusOK,
		"message":     "Login realizado com sucesso!",
		"token":       jwtToken,
		"authTokenID": tokenData.ID,
		"authCode":    tokenData.Code,
		"idUser":      dbUser.ID,
		"name":        dbUser.Name,
	})
}

func (u *UserHandler) validateLoginCredentials(user *model.User, password string) error {
	if err := security.VerifyPasswordMatch(password, user.Password); err != nil {
		return errors.New("invalid login credentials")
	}
	return nil
}

func (u UserHandler) verifyUserActive(user *model.User, c *fiber.Ctx) bool {
	return true
}

func (u UserHandler) handleLoginError(c *fiber.Ctx, errMsg string) error {
	if err := utils.EncodeRequestError(c, errMsg, "/login"); err != nil {
		return fmt.Errorf("failed to encode request error: %s", err)
	}
	return nil
}

func (u *UserHandler) EditUser(c *fiber.Ctx) error {
	var body map[string]interface{}

	if err := c.BodyParser(&body); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "failed to parse body")
	}

	idRaw, ok := body["idUser"]
	if !ok {
		return fiber.NewError(fiber.StatusBadRequest, "idUser é obrigatório")
	}

	userID, ok := idRaw.(string)
	if !ok || userID == "" {
		return fiber.NewError(fiber.StatusBadRequest, "idUser deve ser uma string válida")
	}

	delete(body, "idUser")

	if err := u.UserRepository.UpdateByID(userID, body); err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "falha ao atualizar usuário")
	}

	return c.SendStatus(fiber.StatusOK)
}

type CloudinaryResponse struct {
	SecureURL string `json:"secure_url"`
	PublicID  string `json:"public_id"`
	Error     struct {
		Message string `json:"message"`
	} `json:"error"`
}

func (h *UserHandler) UploadUserImage(c *fiber.Ctx) error {
	userIDStr := c.FormValue("idUser")
	if userIDStr == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "idUser is required",
		})
	}

	user, err := h.UserRepository.FindByFilter("id", userIDStr)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "failed to fetch user",
		})
	}

	if user != nil && user.PhotoUrl != "" {
		publicID := utils.ExtractPublicID(user.PhotoUrl)
		if publicID != "" {
			if err := deleteCloudinaryImage(publicID); err != nil {
				log.Printf("Aviso: não foi possível deletar imagem anterior - %v", err)
			}
		}
	}

	fileHeader, err := c.FormFile("image")
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "no image provided",
		})
	}

	file, err := fileHeader.Open()
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "failed to open image",
		})
	}
	defer file.Close()

	imageBytes, err := io.ReadAll(file)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "failed to read image data",
		})
	}

	if !utils.IsValidImageType(imageBytes) {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "invalid image format. Only JPEG/PNG allowed",
		})
	}

	photoURL, err := utils.UploadPhotoToCloudinary(imageBytes)
	if err != nil {
		log.Printf("Erro no upload para Cloudinary: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "failed to upload image to cloud storage",
		})
	}

	err = h.UserRepository.UpdateByID(userIDStr, map[string]interface{}{
		"photo_url": photoURL,
	})
	if err != nil {
		log.Printf("Erro ao atualizar usuário: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "failed to update user profile",
		})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"message": "Profile photo updated successfully",
	})
}

func deleteCloudinaryImage(publicID string) error {
	cloudName := os.Getenv("CLOUDINARY_CLOUD_NAME")
	apiKey := os.Getenv("CLOUDINARY_API_KEY")
	apiSecret := os.Getenv("CLOUDINARY_API_SECRET")

	timestamp := fmt.Sprintf("%d", time.Now().Unix())
	signatureBase := fmt.Sprintf("public_id=%s&timestamp=%s%s", publicID, timestamp, apiSecret)

	h := sha1.New()
	h.Write([]byte(signatureBase))
	signature := hex.EncodeToString(h.Sum(nil))

	form := url.Values{}
	form.Add("public_id", publicID)
	form.Add("timestamp", timestamp)
	form.Add("api_key", apiKey)
	form.Add("signature", signature)

	deleteURL := fmt.Sprintf("https://api.cloudinary.com/v1_1/%s/image/destroy", cloudName)
	req, err := http.NewRequest("POST", deleteURL, strings.NewReader(form.Encode()))
	if err != nil {
		return err
	}
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	var res map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&res); err != nil {
		return err
	}
	if res["result"] != "ok" {
		return fmt.Errorf("erro ao deletar imagem do cloudinary: %v", res)
	}

	return nil
}

func (h *UserHandler) UploadUserBanner(c *fiber.Ctx) error {
	cloudName := os.Getenv("CLOUDINARY_CLOUD_NAME")
	apiKey := os.Getenv("CLOUDINARY_API_KEY")
	apiSecret := os.Getenv("CLOUDINARY_API_SECRET")

	userIDStr := c.FormValue("idUser")
	if userIDStr == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "idUser is required",
		})
	}

	path := c.FormValue("path")
	if path != "" {
		err := h.UserRepository.UpdateByID(userIDStr, map[string]interface{}{
			"banner": path,
		})
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "failed to update banner path",
			})
		}
		return c.Status(fiber.StatusOK).JSON(fiber.Map{
			"message": "banner atualizado com path local",
		})
	}

	fileHeader, err := c.FormFile("banner")
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "no banner image provided",
		})
	}

	file, err := fileHeader.Open()
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "failed to open banner",
		})
	}
	defer file.Close()

	timestamp := fmt.Sprintf("%d", time.Now().Unix())
	signatureBase := "timestamp=" + timestamp + apiSecret

	hSha := sha1.New()
	hSha.Write([]byte(signatureBase))
	signature := hex.EncodeToString(hSha.Sum(nil))

	var body bytes.Buffer
	writer := multipart.NewWriter(&body)

	_ = writer.WriteField("api_key", apiKey)
	_ = writer.WriteField("timestamp", timestamp)
	_ = writer.WriteField("signature", signature)

	part, err := writer.CreateFormFile("file", fileHeader.Filename)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "error preparing upload",
		})
	}

	_, err = io.Copy(part, file)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "error copying banner file",
		})
	}
	writer.Close()

	uploadURL := fmt.Sprintf("https://api.cloudinary.com/v1_1/%s/image/upload", cloudName)
	req, err := http.NewRequest("POST", uploadURL, &body)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "error creating request",
		})
	}
	req.Header.Set("Content-Type", writer.FormDataContentType())

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "error sending banner",
		})
	}
	defer resp.Body.Close()

	var cloudResp CloudinaryResponse
	if err := json.NewDecoder(resp.Body).Decode(&cloudResp); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "error decoding cloudinary response",
		})
	}

	if cloudResp.SecureURL == "" {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "cloudinary error: " + cloudResp.Error.Message,
		})
	}

	err = h.UserRepository.UpdateByID(userIDStr, map[string]interface{}{
		"banner": cloudResp.SecureURL,
	})
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "failed to update user banner",
		})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"message": "banner enviado com sucesso",
	})
}
