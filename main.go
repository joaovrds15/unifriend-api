package main

import (
	"unifriend-api/models"
	"unifriend-api/routes"

	"github.com/gin-gonic/gin"
)

func main() {
	r := gin.Default()

	// Connect to the database
	models.ConnectDataBase()

	// Configure routes
	routes.SetupRoutes(r)

	// Run the server on port 8090
	r.Run(":8090")
}
