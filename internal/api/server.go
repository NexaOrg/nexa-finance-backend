package api

import (
	"log"
	"nexa/internal/handler"
	"nexa/internal/utils"
	"os"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/jackc/pgx/v5"
	"github.com/joho/godotenv"
)

func SetupRoutes(db *pgx.Conn) {
	_ = godotenv.Load()
	port := os.Getenv("API_PORT")
	app := fiber.New()
	app.Use(cors.New())

	mailServer := utils.NewMailServer(
		os.Getenv("SMTP_SERVER"),
		os.Getenv("SMTP_PORT"),
		os.Getenv("SMTP_FROM"),
		os.Getenv("SMTP_USER"),
		os.Getenv("SMTP_PASSWORD"),
	)

	authHandler := handler.NewUserAuthenticationHandler(db, mailServer)
	userHandler := handler.NewUserHandler(db, authHandler)

	app.Get("/", func(c *fiber.Ctx) error {
		return c.SendString("🚀 Nexa API rodando com sucesso!")
	})

	app.Post("/user", userHandler.RegisterUser)
	app.Post("/auth/login", userHandler.LoginUser)

	log.Printf("Servidor rodando na porta %s", port)
	log.Fatal(app.Listen(":" + port))
}
