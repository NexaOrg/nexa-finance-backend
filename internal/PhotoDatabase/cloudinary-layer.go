package photodatabase

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
	"os"
	"time"
)

func UploadPhotoToCloudinary(imageBytes []byte) (string, error) {
	cloudName := os.Getenv("CLOUDINARY_CLOUD_NAME")
	apiKey := os.Getenv("CLOUDINARY_API_KEY")
	apiSecret := os.Getenv("CLOUDINARY_API_SECRET")

	if cloudName == "" || apiKey == "" || apiSecret == "" {
		return "", fmt.Errorf("cloudinary credentials not found in environment variables")
	}
	if len(imageBytes) == 0 {
		return "", fmt.Errorf("image bytes cannot be empty")
	}

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

	filename := fmt.Sprintf("upload_%d", time.Now().UnixNano())
	part, err := writer.CreateFormFile("file", filename)
	if err != nil {
		return "", fmt.Errorf("error creating form file: %v", err)
	}

	if _, err = part.Write(imageBytes); err != nil {
		return "", fmt.Errorf("error writing image data: %v", err)
	}

	if err = writer.Close(); err != nil {
		return "", fmt.Errorf("error closing writer: %v", err)
	}

	uploadURL := fmt.Sprintf("https://api.cloudinary.com/v1_1/%s/image/upload", cloudName)
	req, err := http.NewRequest("POST", uploadURL, &body)
	if err != nil {
		return "", fmt.Errorf("error creating request: %v", err)
	}
	req.Header.Set("Content-Type", writer.FormDataContentType())

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("upload request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		errorBody, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("cloudinary error [%d]: %s", resp.StatusCode, string(errorBody))
	}

	var result struct {
		SecureURL string `json:"secure_url"`
		Error     struct {
			Message string `json:"message"`
		} `json:"error"`
	}

	if err = json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", fmt.Errorf("error decoding response: %v", err)
	}

	if result.Error.Message != "" {
		return "", fmt.Errorf("cloudinary error: %s", result.Error.Message)
	}

	if result.SecureURL == "" {
		return "", errors.New("empty secure URL in response")
	}

	return result.SecureURL, nil
}
