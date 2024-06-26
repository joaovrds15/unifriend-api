package tests

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"unifriend-api/models"
	"unifriend-api/routes"
	"unifriend-api/tests/factory"

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

	user := factory.UserFactory()
	user.Email = "teste@mail.com"
	user.Password = "wrong"

	models.DB.Create(&user)

	payload := []byte(`{"email": "teste@mail.com", "password": "right"}`)
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

	user := factory.UserFactory()
	user.Email = "test@mail.com"

	user.Password = "right"

	models.DB.Create(&user)
	payload := []byte(`{"email": "test@mail.com", "password": "right"}`)
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
		"password": "senha", 
		"re_password" : "senha",
		"major_id": 1,
		"email": "testemail@mail.com",
		"name": "test user",
		"profile_picture_url": "http://test.com"
	}`)
	req, _ := http.NewRequest("POST", "/api/register", bytes.NewBuffer(payload))
	req.Header.Set("Content-Type", "application/json")

	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)
	models.TearDownTestDB()

	assert.Equal(t, http.StatusCreated, rec.Code)
	assert.Contains(t, rec.Body.String(), "User created successfully")
}

func TestRegisterWithDuplicatedEmail(t *testing.T) {
	router := gin.Default()
	models.SetupTestDB()
	routes.SetupRoutes(router)

	major := factory.MajorFactory()
	user := factory.UserFactory()
	user.Email = "testuser@mail.com"

	models.DB.Create(&major)
	models.DB.Create(&user)

	payload := []byte(`{
		"password": "senha", 
		"re_password" : "senha",
		"major_id": 1,
		"email": "testuser@mail.com",
		"name": "test user",
		"profile_picture_url": "http://test.com"
	}`)
	req, _ := http.NewRequest("POST", "/api/register", bytes.NewBuffer(payload))
	req.Header.Set("Content-Type", "application/json")

	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)
	models.TearDownTestDB()

	assert.Equal(t, http.StatusBadRequest, rec.Code)
	assert.Contains(t, rec.Body.String(), "something went wrong")
}
