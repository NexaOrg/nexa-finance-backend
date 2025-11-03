package database

import (
	"context"
	"fmt"
	"os"

	"github.com/jackc/pgx/v5"
	"github.com/joho/godotenv"
)

func ConnectDB() *pgx.Conn {
	if os.Getenv("RENDER") == "" {
		_ = godotenv.Load()
	}

	user := os.Getenv("DB_USER")
	password := os.Getenv("DB_PASSWORD")
	url := os.Getenv("DB_URL")
	port := os.Getenv("DB_PORT")
	name := os.Getenv("DB_NAME")

	connString := fmt.Sprintf("postgres://%s:%s@%s:%s/%s", user, password, url, port, name)

	conn, err := pgx.Connect(context.Background(), connString)

	if err != nil {
		fmt.Printf("Connection failed: %v", err)
		return nil
	}

	return conn
}
