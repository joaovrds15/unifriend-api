package main

import (
	"unifriend-api/controllers"
	"unifriend-api/middleware"
	"unifriend-api/models"

	"github.com/gin-gonic/gin"
)

func setupRoutes(r *gin.Engine) {
	public := r.Group("/api")
	private := r.Group("/api/private")
	private.Use(middleware.AuthMiddleware())

	private.POST("/answer/save", controllers.SaveAnswers)
	private.GET("/ping", func(ctx *gin.Context) {
		ctx.JSON(200, gin.H{
			"message": "pong",
		})
	})

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
