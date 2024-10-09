package main

import (
	"os"
	"unifriend-api/models"
	"unifriend-api/routes"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

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
		AllowOrigins:     []string{os.Getenv("CLIENT_DOMAIN")},
		AllowMethods:     []string{"POST", "GET", "OPTIONS", "PUT", "DELETE"},
		AllowHeaders:     []string{"Content-Type"},
		AllowCredentials: true,
	}

	r.Use(cors.New(corsConfig))
	routes.SetupRoutes(r)

	r.Run(":8090")
}
