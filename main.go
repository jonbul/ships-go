package main

import (
	"log"
	"os"
	"strings"

	"golang.org/x/sync/errgroup"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"

	"ships/controllers"
	dataaccess "ships/dataAccess"
)

var (
	g errgroup.Group
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

	buildGameWebServer()

	// Start the admin server on a different port
	buildAdminWebServer()

}

func buildGameWebServer() {
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

	g.Go(func() error {
		println("Server is running on port " + port)
		return router.RunTLS(":"+port, os.Getenv("SSL_CERT_PATH"), os.Getenv("SSL_KEY_PATH"))
	})
}

func buildAdminWebServer() {
	adminCorsConfig := cors.DefaultConfig()
	adminCorsConfig.AllowOrigins = []string{"*"}
	adminCorsConfig.AllowCredentials = true
	adminCorsConfig.AddAllowHeaders("*")
	adminCorsConfig.AddAllowMethods("*")

	adminPort := os.Getenv("ADMIN_PORT")
	if adminPort == "" {
		adminPort = "3001"
	}

	adminRouter := gin.Default()
	adminRouter.Use(cors.New(adminCorsConfig))
	controllers.RegisterPrometheusRoutes(adminRouter)

	log.Println("Admin server is running on port " + adminPort)

	g.Go(func() error {
		println("Admin server is running on port " + adminPort)
		return adminRouter.Run(":" + adminPort)
	})

	errG := g.Wait()
	if errG != nil {
		log.Fatal(errG)
	}
}
