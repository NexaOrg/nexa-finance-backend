package main

import (
	"synap/internal/api"
	"synap/internal/database"
)

func main() {
	client := database.MongoConnect()
	api.SetupRoutes(client)
}