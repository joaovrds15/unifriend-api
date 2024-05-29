package main

import (
	"unifriend-api/models"
	"unifriend-api/routes"

	_ "unifriend-api/docs"

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

func main() {
	r := gin.Default()

	// Connect to the database
	models.ConnectDataBase()

	// Configure routes
	routes.SetupRoutes(r)

	// Run the server on port 8090
	r.Run(":8090")
}
