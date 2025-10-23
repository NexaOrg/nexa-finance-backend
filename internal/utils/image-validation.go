package utils

import "net/http"

func IsValidImageType(data []byte) bool {
	contentType := http.DetectContentType(data)
	return contentType == "image/jpeg" || contentType == "image/png"
}
