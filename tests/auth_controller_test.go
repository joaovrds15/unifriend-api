package tests

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"unifriend-api/controllers"
	"unifriend-api/middleware"
	"unifriend-api/models"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
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

func TestLoginWithWrongCredentials(t *testing.T) {
	router := gin.Default()
	models.SetupTestDB()
	setupRoutes(router)

	major := models.Major{
		Name: "Computer Science",
	}

	models.DB.Create(&major)

	user := models.User{
		Username: "testuser",
		Password: "password",
		MajorID:  major.ID,
	}

	models.DB.Create(&user)

	payload := []byte(`{"username": "testuser", "password": "senha"}`)
	req, _ := http.NewRequest("POST", "/api/login", bytes.NewBuffer(payload))
	req.Header.Set("Content-Type", "application/json")

	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)
	models.TearDownTestDB()

	assert.Equal(t, http.StatusUnauthorized, rec.Code)
	assert.Contains(t, rec.Body.String(), "username or password is incorrect.")
}

func TestLogin(t *testing.T) {
	router := gin.Default()
	models.SetupTestDB()
	setupRoutes(router)
	os.Setenv("TOKEN_HOUR_LIFESPAN", "1")

	major := models.Major{
		Name: "Computer Science",
	}

	models.DB.Create(&major)

	user := models.User{
		Username: "testuser",
		Password: "senha",
		MajorID:  major.ID,
	}

	models.DB.Create(&user)
	payload := []byte(`{"username": "testuser", "password": "senha"}`)
	req, _ := http.NewRequest("POST", "/api/login", bytes.NewBuffer(payload))
	req.Header.Set("Content-Type", "application/json")

	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)
	models.TearDownTestDB()

	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Contains(t, rec.Body.String(), "token")
}

func TestRegister(t *testing.T) {
	router := gin.Default()
	models.SetupTestDB()
	setupRoutes(router)

	major := models.Major{
		Name: "Computer Science",
	}

	models.DB.Create(&major)

	payload := []byte(`{
		"username": "testuser",
		"password": "senha", 
		"re_password" : "senha",
		"major_id": 1,
		"email": "testemail@mail.com",
		"first_name": "test",
		"last_name": "user",
		"profile_picture_url": "http://test.com"
	}`)
	req, _ := http.NewRequest("POST", "/api/register", bytes.NewBuffer(payload))
	req.Header.Set("Content-Type", "application/json")

	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)
	models.TearDownTestDB()

	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Contains(t, rec.Body.String(), "registration success")
}
