package routes

import (
	"log"
	"unifriend-api/controllers"
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
	if gin.Mode() != gin.TestMode {
		s3Client, err := services.NewS3Client()
		if err != nil {
			log.Fatalf("Failed to create S3 client: %v", err)
		}

		sesClient, err := services.NewSesClient()
		if err != nil {
			log.Fatalf("Failed to create SES client: %v", err)
		}

		register.POST("/upload-image", func(c *gin.Context) {
			url, err := controllers.UploadImage(c, s3Client)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}
			c.JSON(http.StatusCreated, gin.H{"image-url": url})
		})

		private.PUT("/users/me/profile-picture", func(c *gin.Context) {
			controllers.UpdateUserProfilePicture(c, s3Client)
		})

		private.POST("/users/me/images", func(c *gin.Context) {
			controllers.AddUserImage(c, s3Client)
		})

		private.DELETE("/users/me/images/:image_id", func(c *gin.Context) {
			controllers.DeleteUserImage(c)
		})

		public.GET("/verify/email/:email", func(c *gin.Context) {
			controllers.VerifyEmail(c, sesClient)
		})
	}

	public.GET("/verify/code/:email", controllers.GetVerificationCodeExpiration)
	public.POST("/verify/email", controllers.VerifyEmailCode)
	register.POST("/register", controllers.Register)
	private.POST("/answer/save", controllers.SaveAnswers)
	private.GET("/questions", controllers.GetQuestions)
	private.GET("/get-results/user/:user_id", controllers.GetResults)
	public.POST("/login", controllers.Login)
	public.GET("/health", Ping)
	public.GET("/majors", controllers.GetMajors)
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
