package tests

import (
	"bytes"
	"context"
	"io"
	"log"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"os"
	"strconv"
	"testing"
	"unifriend-api/models"
	"unifriend-api/routes"
	"unifriend-api/tests/factory"

	"github.com/aws/aws-sdk-go-v2/feature/s3/manager"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type MockS3Client struct {
	mock.Mock
}

func (m *MockS3Client) Upload(ctx context.Context, input *s3.GetObjectInput, opts ...func(*manager.Uploader)) (*manager.UploadOutput, error) {
	args := m.Called(ctx, input)
	return args.Get(0).(*manager.UploadOutput), args.Error(1)
}

func createFileForTesting(extension string, size int) *os.File {
	fileName := "image." + extension
	fileSize := int64(size)

	file, err := os.Create(fileName)

	if err != nil {
		log.Fatalf("Failed to create file: %v", err)
	}

	defer file.Close()

	// Set the file size
	if err := file.Truncate(fileSize); err != nil {
		log.Fatalf("Failed to set file size: %v", err)
	}

	return file
}

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
	req, err := http.NewRequest("POST", "/api/upload-image", &buffer)
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}

	req.Header.Set("Content-Type", writer.FormDataContentType())
	req.Header.Set("Authorization", "Bearer "+factory.GetUserFactoryToken(user.ID))

	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusBadRequest, rec.Code)
	assert.Contains(t, rec.Body.String(), "file and user_id are required")
}

func TestUploadImageWithoutUserId(t *testing.T) {
	router := gin.Default()
	models.SetupTestDB()
	routes.SetupRoutes(router)

	user := factory.UserFactory()
	models.DB.Create(user)

	var buffer bytes.Buffer
	writer := multipart.NewWriter(&buffer)

	part, _ := writer.CreateFormFile("file", "image_test.jpg")
	io.Copy(part, createFileForTesting("jpeg", 1024))

	writer.Close()
	req, err := http.NewRequest("POST", "/api/upload-image", &buffer)
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}

	req.Header.Set("Content-Type", writer.FormDataContentType())
	req.Header.Set("Authorization", "Bearer "+factory.GetUserFactoryToken(user.ID))

	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)
	assert.Equal(t, http.StatusBadRequest, rec.Code)
	assert.Contains(t, rec.Body.String(), "file and user_id are required")
}

func TestUploadImageBiggerThenLimit(t *testing.T) {
	router := gin.Default()
	models.SetupTestDB()
	routes.SetupRoutes(router)

	user := factory.UserFactory()
	models.DB.Create(user)

	var buffer bytes.Buffer
	writer := multipart.NewWriter(&buffer)
	_ = writer.WriteField("user_id", strconv.FormatUint(uint64(user.ID), 10))

	image := createFileForTesting("jpg", 1024*1024)
	image.Close()
	imageTest, err := os.Open("image.jpg")
	if err != nil {
		t.Fatalf("Failed to open image file: %v", err)
	}

	defer imageTest.Close()

	part, _ := writer.CreateFormFile("file", "image_test.jpg")
	io.Copy(part, imageTest)

	writer.Close()

	req, err := http.NewRequest("POST", "/api/upload-image", &buffer)
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}

	req.Header.Set("Content-Type", writer.FormDataContentType())
	req.Header.Set("Authorization", "Bearer "+factory.GetUserFactoryToken(user.ID))

	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	os.Remove("image.jpg")

	assert.Equal(t, http.StatusBadRequest, rec.Code)
	assert.Contains(t, rec.Body.String(), "could not upload the file")
}

func TestUploadImageWrongExtension(t *testing.T) {
	router := gin.Default()
	models.SetupTestDB()
	routes.SetupRoutes(router)

	user := factory.UserFactory()
	models.DB.Create(user)

	var buffer bytes.Buffer
	writer := multipart.NewWriter(&buffer)
	_ = writer.WriteField("user_id", strconv.FormatUint(uint64(user.ID), 10))

	image := createFileForTesting("txt", 1024)
	image.Close()
	imageTest, err := os.Open("image.txt")
	if err != nil {
		t.Fatalf("Failed to open image file: %v", err)
	}

	defer imageTest.Close()

	part, _ := writer.CreateFormFile("file", "image_test.txt")
	io.Copy(part, imageTest)

	writer.Close()

	req, err := http.NewRequest("POST", "/api/upload-image", &buffer)
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}

	req.Header.Set("Content-Type", writer.FormDataContentType())
	req.Header.Set("Authorization", "Bearer "+factory.GetUserFactoryToken(user.ID))

	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	os.Remove("image.txt")

	assert.Equal(t, http.StatusBadRequest, rec.Code)
	assert.Contains(t, rec.Body.String(), "could not upload the file")
}

func TestUploadImage(t *testing.T) {
	router := gin.Default()
	models.SetupTestDB()
	routes.SetupRoutes(router)
	os.Setenv("MAX_SIZE_PROFILE_IMAGE_KB", "1024")

	mockUploader := new(MockS3Client)

	mockUploader.On("Upload", mock.Anything, mock.Anything).Return(&manager.UploadOutput{
		Location: "mockLocation",
	}, nil)

	user := factory.UserFactory()
	models.DB.Create(user)

	var buffer bytes.Buffer
	writer := multipart.NewWriter(&buffer)
	_ = writer.WriteField("user_id", strconv.FormatUint(uint64(user.ID), 10))

	image := createFileForTesting("jpg", 1024)
	image.Close()
	imageTest, err := os.Open("image.jpg")
	if err != nil {
		t.Fatalf("Failed to open image file: %v", err)
	}

	defer imageTest.Close()

	part, _ := writer.CreateFormFile("file", "image_test.jpg")
	io.Copy(part, imageTest)

	writer.Close()

	req, err := http.NewRequest("POST", "/api/upload-image", &buffer)
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}

	req.Header.Set("Content-Type", writer.FormDataContentType())
	req.Header.Set("Authorization", "Bearer "+factory.GetUserFactoryToken(user.ID))

	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	os.Remove("image.jpg")

	assert.Equal(t, http.StatusBadRequest, rec.Code)
	assert.Contains(t, rec.Body.String(), "could not upload the file")
}

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
