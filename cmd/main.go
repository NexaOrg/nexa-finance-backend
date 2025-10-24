package main

import (
	"fmt"
	"nexa/internal/api"
	"nexa/internal/database"
)

func main() {
	db := database.ConnectDB()
	if db != nil {
		fmt.Println("Successful connection!")
	}

	api.SetupRoutes(db)
}
