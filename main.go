package main

import (
	"log"
	"os"
	"strings"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"

	"ships/controllers"
	dataaccess "ships/dataAccess"
)

func init() {
	log.Default()
	_ = godotenv.Load()
}

func main() {
	err := godotenv.Load()

	if err != nil {
		log.Fatal("Error loading .env file")
	}

	dataaccess.Test()

	router := gin.Default()
	corsConfig := cors.DefaultConfig()
	corsConfig.AllowOrigins = strings.Split(os.Getenv("ALLOWED_ORIGINS"), "|")
	corsConfig.AllowCredentials = true
	corsConfig.AddAllowHeaders("Authorization")
	corsConfig.AddAllowMethods("GET", "POST", "PUT", "PATCH", "OPTIONS")
	router.Use(cors.New(corsConfig))
	controllers.RegisterRoutes(router)
	port := os.Getenv("PORT")
	if port == "" {
		port = "3000"
	}

	log.Println("Server is running on port " + port)

	err = router.RunTLS(":"+port, "./ssl/cert.pem", "./ssl/key.pem")
	if err != nil {
		log.Fatal("Error setting up SSH Server", err)
	}

	_ = router.Run(":" + port)

}
