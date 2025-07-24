package main

import (
	"log"
	"os"
	"unifriend-api/models"
	"unifriend-api/routes"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

//	@title			UniFriend API
//	@version		1.0
//	@description	API for UniFriend application
//	@termsOfService	http://swagger.io/terms/

//	@contact.name	API Support
//	@contact.url	http://www.swagger.io/support
//	@contact.email	support@swagger.io

//	@license.name	Apache 2.0
//	@license.url	http://www.apache.org/licenses/LICENSE-2.0.html

//	@host		localhost:8090
//	@BasePath	/api

//	@SecurityDefinitions.apiKey	Bearer
//	@in							header
//	@name						Authorization

func main() {
	r := gin.Default()

	err := godotenv.Load(".env")

	if err != nil {
		log.Fatalf("Error loading .env file")
	}

	models.ConnectDataBase()

	corsConfig := cors.Config{
		AllowOrigins:     []string{os.Getenv("CLIENT_DOMAIN")},
		AllowMethods:     []string{"POST", "GET", "OPTIONS", "PUT", "DELETE"},
		AllowHeaders:     []string{"Content-Type"},
		AllowCredentials: true,
	}

	r.Use(cors.New(corsConfig))
	routes.SetupRoutes(r)

	r.Run(":8090")
}
