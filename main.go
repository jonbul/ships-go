package main

import (
	"log"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"

	"ships/controllers"
	dataaccess "ships/dataAccess"
)

func init() {
	log.Default()
}

func main() {
	err := godotenv.Load()

	if err != nil {
		log.Fatal("Error loading .env file")
	}

	dataaccess.Test()

	//envFile, err := godotenv.Read(".env")
	router := gin.Default()
	controllers.RegisterUserRoutes(router)
	router.Run(":3000")
	log.Println("Server is running on port 3000")
}
