package tests

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"
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

func TestLogin(t *testing.T) {
	router := gin.Default()
	models.SetupTestDB()
	setupRoutes(router)
	// Create a new HTTP request with JSON payload
	payload := []byte(`{"username": "testuser", "password": "testpassword"}`)
	req, _ := http.NewRequest("POST", "/api/login", bytes.NewBuffer(payload))
	req.Header.Set("Content-Type", "application/json")

	// Create a new HTTP recorder to capture the response
	rec := httptest.NewRecorder()

	// Serve the request and record the response
	router.ServeHTTP(rec, req)

	// Check the response status code
	if rec.Code != http.StatusOK {
		t.Errorf("Expected status code %d, but got %d", http.StatusOK, rec.Code)
	}

	// Check the response body
	expectedBody := `{"token": "your_token"}`
	if rec.Body.String() != expectedBody {
		t.Errorf("Expected response body %s, but got %s", expectedBody, rec.Body.String())
	}
}
