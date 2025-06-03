package tests

import (
	"os"
	"testing"
	"unifriend-api/controllers"
	"unifriend-api/models"
	"unifriend-api/routes"

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
	routes.SetupRoutes(router)
}

func SetupRouterWithoutMiddleware() *gin.Engine {
	r := gin.Default()

	r.POST("/api/register", controllers.Register)
	return r
}
