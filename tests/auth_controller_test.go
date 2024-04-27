package tests

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"unifriend-api/models"
	"unifriend-api/routes"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestLoginWithWrongCredentials(t *testing.T) {
	router := gin.Default()
	models.SetupTestDB()
	routes.SetupRoutes(router)

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
	routes.SetupRoutes(router)
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
	routes.SetupRoutes(router)

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
