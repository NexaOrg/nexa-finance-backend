package utils

import "strings"

func FormatName(name string) string {
	words := strings.Fields(name)

	for i, word := range words {
		if len(word) > 0 {
			words[i] = strings.ToUpper(string(word[0])) + strings.ToLower(word[1:])
		}
	}

	return name
}
