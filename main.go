package main

import (
	"time"
	"unifriend-api/models"
	"unifriend-api/routes"

	_ "unifriend-api/docs"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
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

	models.ConnectDataBase()
	corsConfig := cors.Config{
		AllowOrigins:     []string{"34.95.247.72"}, // Replace with your frontend URL
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}
	r.Use(cors.New(corsConfig))
	routes.SetupRoutes(r)

	r.Run(":8090")
}
