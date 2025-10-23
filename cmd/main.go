package main

import (
	"fmt"
	"nexa/internal/database"
)

func main() {
	conn := database.ConnectDB()
	if conn != nil {
		fmt.Println("Successful connection!")
	}
}
