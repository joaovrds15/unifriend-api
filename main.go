package main

import (
	"unifriend-api/controllers"
	"unifriend-api/models"

	"github.com/gin-gonic/gin"
)

func setupRoutes(r *gin.Engine) {
	public := r.Group("/api")

	public.POST("/register", controllers.Register)
	public.POST("/login", controllers.Login)
	public.GET("/health", func(ctx *gin.Context) {
		ctx.JSON(200, gin.H{
			"message": "OK",
		})
	})
}

func main() {
	r := gin.Default()

	// Connect to the database
	models.ConnectDataBase()

	// Configure routes
	setupRoutes(r)

	// Run the server on port 8090
	r.Run(":8090")
}
