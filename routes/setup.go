package routes

import (
	"unifriend-api/controllers"
	"unifriend-api/middleware"

	"github.com/gin-gonic/gin"
)

func SetupRoutes(r *gin.Engine) {
	public := r.Group("/api")
	private := r.Group("/api/private")
	private.Use(middleware.AuthMiddleware())

	private.POST("/answer/save", controllers.SaveAnswers)
	private.GET("/ping", func(ctx *gin.Context) {
		ctx.JSON(200, gin.H{
			"message": "pong",
		})
	})

	public.GET("/questions", controllers.GetQuestions)
	public.POST("/register", controllers.Register)
	public.POST("/login", controllers.Login)
	public.GET("/health", func(ctx *gin.Context) {
		ctx.JSON(200, gin.H{
			"message": "OK",
		})
	})
}
