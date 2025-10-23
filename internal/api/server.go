package api

func SetupRoutes(a string) string {
	return a
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
