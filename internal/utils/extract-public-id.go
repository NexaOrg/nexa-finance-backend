package utils

import (
	"path"
	"strings"
)

func ExtractPublicID(url string) string {
	parts := strings.Split(url, "/")
	if len(parts) == 0 {
		return ""
	}
	filename := parts[len(parts)-1]
	return strings.TrimSuffix(filename, path.Ext(filename))
}
