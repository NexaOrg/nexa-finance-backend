package utils

import (
	"os"
)

func InitMailServer() *MailServer {
	server := os.Getenv("SMTP_SERVER")
	port := os.Getenv("SMTP_PORT")
	user := os.Getenv("SMTP_USER")
	password := os.Getenv("SMTP_PASSWORD")
	from := os.Getenv("SMTP_FROM")

	return NewMailServer(server, port, from, user, password)
}
