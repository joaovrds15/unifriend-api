package routes

import (
	"unifriend-api/controllers"
	"unifriend-api/middleware"

	"github.com/gin-gonic/gin"
	swaggerfiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

func SetupRoutes(r *gin.Engine) {
	public := r.Group("/api")
	private := r.Group("/api")
	private.Use(middleware.AuthMiddleware())

	private.POST("/answer/save", controllers.SaveAnswers)
	private.GET("/questions", controllers.GetQuestions)

	public.POST("/register", controllers.Register)
	public.POST("/login", controllers.Login)
	public.GET("/health", Ping)
	private.GET("/majors", controllers.GetMajors)
	public.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerfiles.Handler))
}

// PingExample godoc
//
//	@Summary	ping route
//	@Schemes
//	@Description	do ping
//	@Tags			health
//	@Accept			json
//	@Produce		json
//	@Success		200	{string}	pong
//	@Router			/health [get]
func Ping(ctx *gin.Context) {
	ctx.JSON(200, gin.H{
		"message": "pong",
	})
}
