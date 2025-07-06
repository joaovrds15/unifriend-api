package routes

import (
	"log"
	"unifriend-api/handlers"
	"unifriend-api/middleware"
	"unifriend-api/services"

	"github.com/gin-gonic/gin"
	swaggerfiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

func SetupRoutes(r *gin.Engine) {
	public := r.Group("/api")
	private := r.Group("/api")
	register := r.Group("/api")

	private.Use(middleware.AuthMiddleware())
	register.Use(middleware.AuthRegistrationMiddleware())
	
	users := private.Group("/users")
	connections := private.Group("/connections")

	if gin.Mode() != gin.TestMode {
		s3Client, err := services.NewS3Client()
		if err != nil {
			log.Fatalf("Failed to create S3 client: %v", err)
		}

		sesClient, err := services.NewSesClient()
		if err != nil {
			log.Fatalf("Failed to create SES client: %v", err)
		}

		users.POST("/profile-picture", func(c *gin.Context) {
			handlers.UpdateUserProfilePicture(c, s3Client)
		})

		users.PUT("/profile-picture", func(c *gin.Context) {
			handlers.UpdateUserProfilePicture(c, s3Client)
		})

		users.DELETE("/profile-picture", func(c *gin.Context) {
			handlers.DeleteUserProfilePicture(c, s3Client)
		})

		users.POST("/images", func(c *gin.Context) {
			handlers.AddUserImage(c, s3Client)
		})

		users.DELETE("/images/:image_id", func(c *gin.Context) {
			handlers.DeleteUserImage(c, s3Client)
		})

		public.GET("/verify/email/:email", func(c *gin.Context) {
			handlers.VerifyEmail(c, sesClient)
		})
	}
	connections.POST("/request/user/:user_id", handlers.CreateConnectionRequest)
	connections.GET("/requests", handlers.GetConnectionRequests)
	connections.PUT("/requests/:request_id/accept", handlers.AcceptConnectionRequest)
	connections.PUT("/requests/:request_id/reject", handlers.RejectConnectionRequest)
	connections.DELETE("/:connection_id", handlers.DeleteConnection)
	public.GET("/verify/code/:email", handlers.GetVerificationCodeExpiration)
	private.GET("/questions", handlers.GetQuestions)
	private.GET("/get-results/user/:user_id", handlers.GetResults)
	public.GET("/health", Ping)
	public.GET("/majors", handlers.GetMajors)
	public.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerfiles.Handler))
	public.POST("/verify/email", handlers.VerifyEmailCode)
	register.POST("/register", handlers.Register)
	private.POST("/answer/save", handlers.SaveAnswers)
	private.GET("/logout", handlers.Logout)
	public.POST("/login", handlers.Login)
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
