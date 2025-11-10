package api

import (
	"log"
	"nexa/internal/handler"
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

	userHandler := handler.NewUserHandler(db)

	app.Get("/", func(c *fiber.Ctx) error {
		return c.SendString("🚀 Nexa API rodando com sucesso!")
	})

	app.Post("/user", userHandler.RegisterUser)

	log.Printf("Servidor rodando na porta %s", port)
	log.Fatal(app.Listen(":" + port))
}

// func SetupRoutes(client *mongo.Client) {
// 	_ = godotenv.Load()
// 	port := os.Getenv("API_PORT")

// 	app := fiber.New()
// 	app.Use(cors.New())

// 	mailServer := utils.InitMailServer()
// 	authHandler := handler.NewUserAuthenticationHandler(client, mailServer)
// 	userHandler := handler.NewUserHandler(client, authHandler)
// 	followHandler := handler.NewFollowHandler(client)
// 	projectHandler := handler.NewProjectHandler(client)

// 	app.Post("/user", userHandler.RegisterUser)

// 	log.Printf("Servidor rodando na porta %s", port)
// 	log.Fatal(app.Listen(fmt.Sprintf(":%s", port)))
// }
