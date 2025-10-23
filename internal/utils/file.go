package utils

import (
	"fmt"
	"path/filepath"
	"text/template"
)

func ParseFile(filePath string) (*template.Template, error) {
	path, err := filepath.Abs(filePath)
	if err != nil {
		return nil, fmt.Errorf("error getting file path: %s", err)
	}

	t, err := template.ParseFiles(path)
	if err != nil {
		return nil, fmt.Errorf("error parsing file: %s", err)
	}

	return t, nil
}
