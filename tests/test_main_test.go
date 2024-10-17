package tests

import (
	"os"
	"testing"
	"unifriend-api/models"
	"unifriend-api/routes"

	"github.com/gin-gonic/gin"
)

var router *gin.Engine

func TestMain(m *testing.M) {

	router = gin.Default()
	gin.SetMode(gin.TestMode)
	routes.SetupRoutes(router)

	code := m.Run()

	os.Exit(code)
}

func SetupTestDB() {
	models.SetupTestDB()
}
