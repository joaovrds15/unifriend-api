package tests

import (
	"bytes"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"os"
	"strconv"
	"testing"
	"unifriend-api/controllers"
	"unifriend-api/models"
	"unifriend-api/routes"
	"unifriend-api/tests/factory"
	"unifriend-api/tests/mocks"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestUploadImageWithoutImage(t *testing.T) {
	router := gin.Default()
	models.SetupTestDB()
	routes.SetupRoutes(router)

	user := factory.UserFactory()
	models.DB.Create(&user)

	var buffer bytes.Buffer
	writer := multipart.NewWriter(&buffer)

	_ = writer.WriteField("user_id", strconv.FormatUint(uint64(user.ID), 10))
	writer.Close()
	req, err := http.NewRequest("POST", "/api/upload-profile-image", &buffer)
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}

	req.Header.Set("Content-Type", writer.FormDataContentType())
	authCookie := &http.Cookie{
		Name:  "auth_token",
		Value: factory.GetUserFactoryToken(user.ID),
		Path:  "/",
	}

	req.AddCookie(authCookie)

	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusBadRequest, rec.Code)
	assert.Contains(t, rec.Body.String(), "file and user_id are required")
}

func TestUploadImageInvalidExtension(t *testing.T) {
	models.SetupTestDB()
	os.Setenv("MAX_SIZE_PROFILE_IMAGE_KB", "512")

	mockUploader := &mocks.MockS3Uploader{
		UploadImageFunc: func(file multipart.File, fileName string) (string, error) {
			return "https:s3.com.br", nil
		},
	}

	router := gin.Default()

	router.POST("/api/upload-profile-image", func(c *gin.Context) {
		controllers.UploadProfileImage(c, mockUploader)
	})

	user := factory.UserFactory()

	models.DB.Create(&user)

	var body bytes.Buffer
	writer := multipart.NewWriter(&body)

	part, err := writer.CreateFormFile("file", "test.txt")

	if err != nil {
		t.Fatalf("Failed to create form file: %v", err)
	}

	part.Write([]byte("fake image data"))

	_ = writer.WriteField("user_id", strconv.FormatUint(uint64(user.ID), 10))

	writer.Close()

	req, err := http.NewRequest("POST", "/api/upload-profile-image", &body)

	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}

	req.Header.Set("Content-Type", writer.FormDataContentType())
	authCookie := &http.Cookie{
		Name:  "auth_token",
		Value: factory.GetUserFactoryToken(user.ID),
		Path:  "/",
	}

	req.AddCookie(authCookie)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusBadRequest, rec.Code)
	assert.Contains(t, rec.Body.String(), "could not upload the file")
}

func TestUploadImageSuccess(t *testing.T) {
	models.SetupTestDB()
	os.Setenv("MAX_SIZE_PROFILE_IMAGE_KB", "512")

	mockUploader := &mocks.MockS3Uploader{
		UploadImageFunc: func(file multipart.File, fileName string) (string, error) {
			return "https:s3.com.br", nil
		},
	}

	router := gin.Default()

	router.POST("/api/upload-profile-image", func(c *gin.Context) {
		controllers.UploadProfileImage(c, mockUploader)
	})

	user := factory.UserFactory()

	models.DB.Create(&user)

	var body bytes.Buffer
	writer := multipart.NewWriter(&body)

	part, err := writer.CreateFormFile("file", "test.jpg")

	if err != nil {
		t.Fatalf("Failed to create form file: %v", err)
	}

	part.Write([]byte("fake image data"))

	_ = writer.WriteField("user_id", strconv.FormatUint(uint64(user.ID), 10))

	writer.Close()

	req, err := http.NewRequest("POST", "/api/upload-profile-image", &body)

	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}

	req.Header.Set("Content-Type", writer.FormDataContentType())
	authCookie := &http.Cookie{
		Name:  "auth_token",
		Value: factory.GetUserFactoryToken(user.ID),
		Path:  "/",
	}

	req.AddCookie(authCookie)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusCreated, rec.Code)
	assert.Contains(t, rec.Body.String(), "image uploaded succesufuly")
}

func TestUploadImageBiggerThenLimit(t *testing.T) {
	models.SetupTestDB()
	os.Setenv("MAX_SIZE_PROFILE_IMAGE_KB", "512")

	mockUploader := &mocks.MockS3Uploader{
		UploadImageFunc: func(file multipart.File, fileName string) (string, error) {
			return "https:s3.com.br", nil
		},
	}

	router := gin.Default()

	router.POST("/api/upload-profile-image", func(c *gin.Context) {
		controllers.UploadProfileImage(c, mockUploader)
	})

	user := factory.UserFactory()

	models.DB.Create(&user)

	largeFile := bytes.Repeat([]byte("A"), 1<<20)
	var body bytes.Buffer
	writer := multipart.NewWriter(&body)

	part, err := writer.CreateFormFile("file", "test.jpg")

	if err != nil {
		t.Fatalf("Failed to create form file: %v", err)
	}

	part.Write(largeFile)

	_ = writer.WriteField("user_id", strconv.FormatUint(uint64(user.ID), 10))

	writer.Close()

	req, err := http.NewRequest("POST", "/api/upload-profile-image", &body)

	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}

	req.Header.Set("Content-Type", writer.FormDataContentType())
	authCookie := &http.Cookie{
		Name:  "auth_token",
		Value: factory.GetUserFactoryToken(user.ID),
		Path:  "/",
	}

	req.AddCookie(authCookie)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusBadRequest, rec.Code)
	assert.Contains(t, rec.Body.String(), "could not upload the file")
}

func TestLoginWithWrongCredentials(t *testing.T) {
	SetupTestDB()
	defer models.TearDownTestDB()

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

	assert.Equal(t, http.StatusUnauthorized, rec.Code)
	assert.Contains(t, rec.Body.String(), "username or password is incorrect.")
}

func TestLogin(t *testing.T) {
	SetupTestDB()
	defer models.TearDownTestDB()

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

	assert.Equal(t, http.StatusNoContent, rec.Code)
	assert.NotEmpty(t, rec.Header().Get("Set-Cookie"))
}

func TestRegister(t *testing.T) {
	SetupTestDB()
	defer models.TearDownTestDB()

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

	assert.Equal(t, http.StatusCreated, rec.Code)
	assert.Contains(t, rec.Body.String(), "User created successfully")
}
func TestRegisterWithDuplicatedEmail(t *testing.T) {
	SetupTestDB()
	defer models.TearDownTestDB()

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

	assert.Equal(t, http.StatusBadRequest, rec.Code)
	assert.Contains(t, rec.Body.String(), "something went wrong")
}
