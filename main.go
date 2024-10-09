package main

import (
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

	// Custom CORS configuration
	corsConfig := cors.Config{
		AllowOrigins:     []string{"http://localhost:3000"},
		AllowMethods:     []string{"POST", "GET"},
		AllowHeaders:     []string{"Content-Type"},
		AllowCredentials: true,
	}

	r.Use(cors.New(corsConfig))
	routes.SetupRoutes(r)

	r.Run(":8090")
}
