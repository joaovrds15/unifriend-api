package tests

import (
	"os"
	"testing"
	"unifriend-api/handlers"
	"unifriend-api/models"
	"unifriend-api/routes"
	"unifriend-api/services"

	"github.com/gin-gonic/gin"
)

var router *gin.Engine

func TestMain(m *testing.M) {	
	gin.SetMode(gin.TestMode)
	SetupRoutes()
	code := m.Run()

	os.Exit(code)
}

func SetupTestDB() {
	models.SetupTestDB()
}

func SetupRoutes() {
	router = gin.Default()
	hub := services.NewHub()
    routes.SetupRoutes(router, hub)
}

func SetupRouterWithoutMiddleware() *gin.Engine {
	r := gin.Default()

	r.POST("/api/register", handlers.Register)
	return r
}
