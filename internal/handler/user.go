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

	"github.com/gofiber/fiber/v2"
	"github.com/jackc/pgx/v5"
	"github.com/rs/zerolog/log"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type UserHandler struct {
	UserFactory    *factory.UserFactory
	UserRepository *repository.UserRepository
}

func NewUserHandler(db *pgx.Conn) *UserHandler {
	return &UserHandler{
		UserRepository: repository.NewUserRepository(db),
		UserFactory:    factory.NewUserFactory(),
	}
}

func (u *UserHandler) RegisterUser(c *fiber.Ctx) error {
	var model model.User

	if err := c.BodyParser(&model); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Dados inválidos"})
	}

	model.LastLogin = time.Now()

	err := u.UserRepository.InsertUser(&model)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"message": "Internal Server Error", "error": err.Error()})
	}

	return c.Status(201).JSON(fiber.Map{"message": "User creation successful"})
}

func (u *UserHandler) validateUser(user *model.User) (bool, error) {
	// if user.Name == "" || user.Email == "" || user.Password == "" {
	// 	return false, fmt.Errorf("todos os campos são obrigatórios")
	// }

	// if _, err := u.validateName(user.Name); err != nil {
	// 	return false, err
	// }

	// _, _, err := u.validateEmail(user.Email)
	return false, nil
}

func (u *UserHandler) validateEmail(email string) (*primitive.ObjectID, string, error) {
	// if !utils.IsEmailValid(email) {
	// 	return nil, "INVALID_EMAIL", fmt.Errorf("formato de e-mail inválido")
	// }

	// user, err := u.UserRepository.FindByFilter("email", email)
	// if err != nil {
	// 	fmt.Println(err)
	// 	return nil, "INTERNAL_SERVER_ERROR", fmt.Errorf("erro ao acessar o banco de dados")
	// }

	// if user == nil {
	// 	return nil, "", nil
	// }

	// if user.IsActive {
	// 	return nil, "USER_ALREADY_REGISTERED", fmt.Errorf("e-mail já registrado")
	// }

	return nil, "", nil
}

func (u *UserHandler) validateName(name string) (string, error) {
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
		return fmt.Errorf("failed to parse body: %s", err)
	}

	email := strings.ToLower(user.Email)

	dbUser, err := u.UserRepository.FindByFilter("email", strings.TrimSpace(email))
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			if err := utils.EncodeRequestError(c, "INVALID_LOGIN_CREDENTIALS", "/login"); err != nil {
				log.Error().Err(err).Msg("failed to send JSON response")

				return fmt.Errorf("failed to encode request error: %s", err)
			}

			return nil
		}

		if err := utils.EncodeRequestError(c, err.Error(), "/login"); err != nil {
			log.Error().Err(err).Msg("failed to send JSON response")

			return fmt.Errorf("failed to encode request error: %s", err)
		}

		return nil
	}

	if err := u.validateLoginCredentials(dbUser, user.Password); err != nil {
		if err := utils.EncodeRequestError(c, "INVALID_LOGIN_CREDENTIALS", "/login"); err != nil {
			log.Error().Err(err).Msg("failed to send JSON response")

			return fmt.Errorf("failed to encode request error: %s", err)
		}

		return nil
	}

	// if !dbUser.IsActive {
	// 	if err := c.Status(400).JSON(fiber.Map{
	// 		"status":  400,
	// 		"message": "This user is not active.",
	// 		"idUser":  dbUser.IDUser,
	// 	}); err != nil {
	// 		log.Error().Err(err).Msg("failed to send JSON response")

	// 		return fmt.Errorf("failed to encode response: %s", err)
	// 	}

	// 	return nil
	// }

	if !u.verifyUserActive(dbUser, c) {
		return nil
	}

	// token, err := u.UserAuthenticationHandler.CreateToken(dbUser.IDUser, "/login")

	if err != nil {
		if err := utils.EncodeRequestError(c, "INTERNAL_SERVER_ERROR", "/login"); err != nil {
			log.Error().Err(err).Msg("failed to send JSON response")

			return fmt.Errorf("failed to encode request error: %s", err)
		}

		return nil
	}
	if err := c.Status(fiber.StatusOK).JSON(fiber.Map{
		"status":  fiber.StatusOK,
		"message": "Login realizado com sucesso!",
		// "token":   token,
		// "idUser":  dbUser.IDUser,
		// "name":    dbUser.Name,
	}); err != nil {
		log.Error().Err(err).Msg("failed to send JSON response")

		return fmt.Errorf("failed to encode response: %s", err)
	}

	return nil
}

func (u *UserHandler) validateLoginCredentials(user *model.User, password string) error {
	if err := security.VerifyPasswordMatch(password, user.Password); err != nil {
		return errors.New("invalid login credentials")
	}

	return nil
}

func (u UserHandler) verifyUserActive(user *model.User, c *fiber.Ctx) bool {
	// if !user.IsActive {
	// 	_ = u.handleLoginError(c, "USER_NOT_ACTIVE")
	// 	return false
	// }

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
		return fmt.Errorf("failed to parse body: %w", err)
	}

	idRaw, ok := body["idUser"]
	if !ok {
		return fiber.NewError(fiber.StatusBadRequest, "idUser é obrigatório")
	}

	idStr, ok := idRaw.(string)
	if !ok {
		return fiber.NewError(fiber.StatusBadRequest, "idUser deve ser uma string")
	}
	userID, err := primitive.ObjectIDFromHex(idStr)
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "idUser inválido")
	}

	delete(body, "idUser")

	if err := u.UserRepository.UpdateByID(userID, body); err != nil {
		return fmt.Errorf("falha ao atualizar usuário: %w", err)
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

	userID, err := primitive.ObjectIDFromHex(userIDStr)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "invalid user id format",
		})
	}

	user, err := h.UserRepository.FindByFilter("_id", userID)
	if err == nil && user.PhotoUrl != "" {
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

	err = h.UserRepository.UpdateByID(userID, map[string]interface{}{
		"photoUrl": photoURL,
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

	userID, err := primitive.ObjectIDFromHex(userIDStr)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "invalid user id format",
		})
	}

	path := c.FormValue("path")
	if path != "" {
		err = h.UserRepository.UpdateByID(userID, map[string]interface{}{
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

	// user, err := h.UserRepository.FindByFilter("_id", userID)
	// if err == nil && user.Banner != "" {
	// 	publicID := utils.ExtractPublicID(user.Banner)
	// 	if publicID != "" {
	// 		_ = deleteCloudinaryImage(publicID)
	// 	}
	// }

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
		return err
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

	err = h.UserRepository.UpdateByID(userID, map[string]interface{}{
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
