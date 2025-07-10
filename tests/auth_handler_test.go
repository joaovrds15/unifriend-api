package tests

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"
	"unifriend-api/models"
	"unifriend-api/tests/factory"

	"github.com/stretchr/testify/assert"
)

func TestLoginWithWrongCredentials(t *testing.T) {
    SetupTestDB()
    defer models.TearDownTestDB()

    user := factory.UserFactory()
    user.Email = "teste@mail.com"
    user.Password = "Wrong@passowrd"

    models.DB.Create(&user)

    payload := []byte(`{"email": "teste@mail.com", "password": "Right@Password"}`)
    req, _ := http.NewRequest("POST", "/api/login", bytes.NewBuffer(payload))
    req.Header.Set("Content-Type", "application/json")

    rec := httptest.NewRecorder()

    router.ServeHTTP(rec, req)

    assert.Equal(t, http.StatusUnauthorized, rec.Code)
    assert.Contains(t, rec.Body.String(), "username or password is incorrect.")
}

func TestLogin(t *testing.T) {
    SetupTestDB()
    defer models.TearDownTestDB()

    os.Setenv("TOKEN_HOUR_LIFESPAN", "1")

    user := factory.UserFactory()
    user.Email = "test@mail.com"
    user.Password = "Right@Password"

    models.DB.Create(&user)
    payload := []byte(`{"email": "test@mail.com", "password": "Right@Password"}`)
    req, _ := http.NewRequest("POST", "/api/login", bytes.NewBuffer(payload))
    req.Header.Set("Content-Type", "application/json")

    rec := httptest.NewRecorder()

    router.ServeHTTP(rec, req)

    assert.Equal(t, http.StatusOK, rec.Code)
    assert.NotEmpty(t, rec.Header().Get("Set-Cookie"))

    var response map[string]interface{}
    err := json.Unmarshal(rec.Body.Bytes(), &response)
    assert.NoError(t, err)

    assert.False(t, response["error"].(bool))
    assert.NotNil(t, response["data"])
    assert.NotEmpty(t, response["token"])

    userData := response["data"].(map[string]interface{})
    assert.Equal(t, float64(user.ID), userData["id"])
    assert.Equal(t, user.Name, userData["name"])
    assert.Equal(t, user.Email, userData["email"])
    assert.Equal(t, user.PhoneNumber, userData["phone_number"])
    assert.Equal(t, user.ProfilePictureURL, userData["profile_picture_url"])

    token := response["token"].(string)
    assert.NotEmpty(t, token)
}

func TestLoginDeletedUser(t *testing.T) {
    SetupTestDB()
    defer models.TearDownTestDB()

    os.Setenv("TOKEN_HOUR_LIFESPAN", "1")

    user := factory.UserFactory()
	user.Status = 0
	user.DeletedAt = time.Now()
    user.Email = "test@mail.com"
    user.Password = "Right@Password"

    models.DB.Create(&user)
    payload := []byte(`{"email": "test@mail.com", "password": "Right@Password"}`)
    req, _ := http.NewRequest("POST", "/api/login", bytes.NewBuffer(payload))
    req.Header.Set("Content-Type", "application/json")

    rec := httptest.NewRecorder()

    router.ServeHTTP(rec, req)

   	assert.Equal(t, http.StatusUnauthorized, rec.Code)
	assert.Contains(t, rec.Body.String(), "username or password is incorrect.")
}

func TestRegister(t *testing.T) {
	SetupTestDB()
	defer models.TearDownTestDB()

	major := factory.MajorFactory()

	models.DB.Create(&major)

	userData := map[string]interface{}{
		"password":            "Senha@123",
		"re_password":         "Senha@123",
		"major_id":            major.ID,
		"email":               "testemail@mail.com",
		"name":                "test user",
		"phone_number":        "62999999999",
	}

	jsonValue, _ := json.Marshal(userData)

	req, _ := http.NewRequest("POST", "/api/register", bytes.NewBuffer(jsonValue))
	req.Header.Set("Content-Type", "application/json")

	registrationCookie := &http.Cookie{
		Name:  "registration_token",
		Value: factory.GetEmailToken(userData["email"].(string)),
		Path:  "/",
	}

	req.AddCookie(registrationCookie)

	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	var user models.User
	models.DB.Where("email = ?", userData["email"]).First(&user)
	assert.NotNil(t, user)

	var userImages []models.UsersImages
	models.DB.Where("user_id = ?", user.ID).Find(&userImages)

	expectedImageUrls := make(map[string]bool)

	for _, dbImg := range userImages {
		assert.True(t, expectedImageUrls[dbImg.ImageUrl])
	}

	assert.Equal(t, http.StatusCreated, rec.Code)
	assert.Contains(t, rec.Body.String(), "User created successfully")
}

func TestRegisterWithDuplicatedEmail(t *testing.T) {
	SetupTestDB()
	defer models.TearDownTestDB()

	user := factory.UserFactory()
	models.DB.Create(&user)

	userData := map[string]interface{}{
		"password":            "Senha@123",
		"re_password":         "Senha@123",
		"major_id":            1,
		"email":               user.Email,
		"name":                "test user",
		"phone_number":        "62999999999",
	}

	jsonValue, _ := json.Marshal(userData)

	req, _ := http.NewRequest("POST", "/api/register", bytes.NewBuffer(jsonValue))
	req.Header.Set("Content-Type", "application/json")

	rec := httptest.NewRecorder()

	router := SetupRouterWithoutMiddleware()
	router.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusBadRequest, rec.Code)
	assert.Contains(t, rec.Body.String(), "something went wrong")
}

func TestRegisterWithDuplicatedPhoneNumber(t *testing.T) {
	SetupTestDB()
	defer models.TearDownTestDB()

	major := factory.MajorFactory()
	user := factory.UserFactory()
	user.PhoneNumber = "62999999999"

	models.DB.Create(&major)
	models.DB.Create(&user)

	imagesUrls := []string{
		"http://test.com",
		"http://test2.com",
	}

	userData := map[string]interface{}{
		"password":            "Senha@123",
		"re_password":         "Senha@123",
		"major_id":            1,
		"email":               "newuser@mail.com",
		"name":                "new user",
		"profile_picture_url": "http://test.com",
		"phone_number":        "62999999999",
		"images":              imagesUrls,
	}

	jsonValue, _ := json.Marshal(userData)

	req, _ := http.NewRequest("POST", "/api/register", bytes.NewBuffer(jsonValue))
	req.Header.Set("Content-Type", "application/json")

	rec := httptest.NewRecorder()

	router := SetupRouterWithoutMiddleware()
	router.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusBadRequest, rec.Code)
	assert.Contains(t, rec.Body.String(), "something went wrong")
}

func TestRegisterInvalidPassword(t *testing.T) {
	SetupTestDB()

	defer models.TearDownTestDB()

	major := models.Major{
		Name: "Computer Science",
	}

	models.DB.Create(&major)

	userData := map[string]interface{}{
		"password":            "Senha123",
		"re_password":         "Senha123",
		"major_id":            major.ID,
		"email":               "testemail@mail.com",
		"name":                "test user",
		"profile_picture_url": "http://test.com",
		"phone_number":        "62999999999",
	}

	jsonValue, _ := json.Marshal(userData)

	req, _ := http.NewRequest("POST", "/api/register", bytes.NewBuffer(jsonValue))
	req.Header.Set("Content-Type", "application/json")

	rec := httptest.NewRecorder()

	router := SetupRouterWithoutMiddleware()
	router.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusBadRequest, rec.Code)
	assert.Contains(t, rec.Body.String(), "Password must be at least 8 characters long, contain at least one uppercase letter, and one special symbol")
}

func TestLogout(t *testing.T) {
    SetupTestDB()
    defer models.TearDownTestDB()

    os.Setenv("TOKEN_HOUR_LIFESPAN", "1")

    user := factory.UserFactory()
    user.Email = "test@mail.com"
    user.Password = "Right@Password"

    models.DB.Create(&user)
    req, _ := http.NewRequest("GET", "/api/logout", nil)

	authCookie := &http.Cookie{
        Name:  "auth_token",
        Value: factory.GetUserFactoryToken(user.ID),
        Path:  "/",
    }

    req.AddCookie(authCookie)	

    rec := httptest.NewRecorder()

    router.ServeHTTP(rec, req)

    assert.Equal(t, http.StatusNoContent, rec.Code)
	assert.Contains(t, rec.Header().Get("Set-Cookie"), "auth_token=;")
}
