package tests

import (
	"bytes"
	"errors"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"os"
	"strconv"
	"testing"
	"time"
	"unifriend-api/controllers"
	"unifriend-api/models"
	"unifriend-api/tests/factory"
	"unifriend-api/tests/mocks"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestUploadImageWithoutImage(t *testing.T) {
	SetupTestDB()
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
	SetupTestDB()
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
	SetupTestDB()

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
	SetupTestDB()

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

	major := models.Major{
		Name: "Computer Science",
	}

	models.DB.Create(&major)

	user := factory.UserFactory()
	user.Email = "test@mail.com"

	user.Password = "Right@Password"

	models.DB.Create(&user)
	payload := []byte(`{"email": "test@mail.com", "password": "Right@Password"}`)
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
		"password": "Senha@123", 
		"re_password" : "Senha@123",
		"major_id": 1,
		"email": "testemail@mail.com",
		"name": "test user",
		"profile_picture_url": "http://test.com",
		"phone_number": "62999999999"
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
		"password": "Senha@123", 
		"re_password" : "Senha@123",
		"major_id": 1,
		"email": "testuser@mail.com",
		"name": "test user",
		"profile_picture_url": "http://test.com",
		"phone_number": "62999999999"
	}`)
	req, _ := http.NewRequest("POST", "/api/register", bytes.NewBuffer(payload))
	req.Header.Set("Content-Type", "application/json")

	rec := httptest.NewRecorder()

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

	payload := []byte(`{
		"password": "Senha@123", 
		"re_password" : "Senha@123",
		"major_id": 1,
		"email": "newuser@mail.com",
		"name": "new user",
		"profile_picture_url": "http://test.com",
		"phone_number": "62999999999"
	}`)
	req, _ := http.NewRequest("POST", "/api/register", bytes.NewBuffer(payload))
	req.Header.Set("Content-Type", "application/json")

	rec := httptest.NewRecorder()

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

	payload := []byte(`{
		"password": "Senha123", 
		"re_password" : "Senha123",
		"major_id": 1,
		"email": "testemail@mail.com",
		"name": "test user",
		"profile_picture_url": "http://test.com",
		"phone_number": "62999999999"
	}`)

	req, _ := http.NewRequest("POST", "/api/register", bytes.NewBuffer(payload))
	req.Header.Set("Content-Type", "application/json")

	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusBadRequest, rec.Code)
	assert.Contains(t, rec.Body.String(), "Password must be at least 8 characters long, contain at least one uppercase letter, and one special symbol")
}

func TestVerifyEmailWithoutEmailParameter(t *testing.T) {
	SetupTestDB()

	mockEmailSender := &mocks.MockSesSender{
		SendVerificationEmailFunc: func(recipient, subject, body string) error {
			return nil
		},
	}

	router := gin.Default()

	router.GET("/api/verify-email", func(c *gin.Context) {
		controllers.VerifyEmail(c, mockEmailSender)
	})

	gin.SetMode(gin.TestMode)
	req, _ := http.NewRequest("GET", "/api/verify-email", nil)
	req.Header.Set("Content-Type", "application/json")

	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)
	assert.Equal(t, http.StatusBadRequest, rec.Code)
	assert.Contains(t, rec.Body.String(), "email parameter is required")
}

func TestVerifyWithInvalidEmail(t *testing.T) {
	SetupTestDB()

	mockEmailSender := &mocks.MockSesSender{
		SendVerificationEmailFunc: func(recipient, subject, body string) error {
			return nil
		},
	}

	router := gin.Default()

	router.GET("/api/verify/email/:email", func(c *gin.Context) {
		controllers.VerifyEmail(c, mockEmailSender)
	})

	gin.SetMode(gin.TestMode)
	req, _ := http.NewRequest("GET", "/api/verify/email/invalidemail.com", nil)
	req.Header.Set("Content-Type", "application/json")

	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	assert.Equal(t, http.StatusBadRequest, rec.Code)
	assert.Contains(t, rec.Body.String(), "invalid email")
}

func TestVerifyWithInvalidEmailDomain(t *testing.T) {
	SetupTestDB()

	emailDomain := factory.EmailDomainsFactory()
	invalidEmail := "email@invalid.com"
	models.DB.Create(&emailDomain)

	mockEmailSender := &mocks.MockSesSender{
		SendVerificationEmailFunc: func(recipient, subject, body string) error {
			return nil
		},
	}

	router := gin.Default()

	router.GET("/api/verify/email/:email", func(c *gin.Context) {
		controllers.VerifyEmail(c, mockEmailSender)
	})

	gin.SetMode(gin.TestMode)
	req, _ := http.NewRequest("GET", "/api/verify/email/"+invalidEmail, nil)
	req.Header.Set("Content-Type", "application/json")

	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	assert.Equal(t, http.StatusBadRequest, rec.Code)
	assert.Contains(t, rec.Body.String(), "invalid email")
}

func TestVerifyEmailAwsError(t *testing.T) {
	SetupTestDB()

	emailDomain := factory.EmailDomainsFactory()
	validEmail := "email@" + emailDomain.Domain
	models.DB.Create(&emailDomain)

	mockEmailSender := &mocks.MockSesSender{
		SendVerificationEmailFunc: func(recipient, subject, body string) error {
			return errors.New("AWS SES error: Email sending failed")
		},
	}

	router := gin.Default()

	router.GET("/api/verify/email/:email", func(c *gin.Context) {
		controllers.VerifyEmail(c, mockEmailSender)
	})

	gin.SetMode(gin.TestMode)
	req, _ := http.NewRequest("GET", "/api/verify/email/"+validEmail, nil)
	req.Header.Set("Content-Type", "application/json")

	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)
	assert.Equal(t, http.StatusInternalServerError, rec.Code)
	assert.Contains(t, rec.Body.String(), "something went wrong")
}

func TestVerifyEmailWithValidCode(t *testing.T) {
	SetupTestDB()

	emailDomain := factory.EmailDomainsFactory()
	validEmail := "email@" + emailDomain.Domain
	var emailVerification models.EmailsVerification
	models.DB.Create(&emailDomain)

	mockEmailSender := &mocks.MockSesSender{
		SendVerificationEmailFunc: func(recipient, subject, body string) error {
			return nil
		},
	}

	router := gin.Default()

	router.GET("/api/verify/email/:email", func(c *gin.Context) {
		controllers.VerifyEmail(c, mockEmailSender)
	})

	gin.SetMode(gin.TestMode)
	req, _ := http.NewRequest("GET", "/api/verify/email/"+validEmail, nil)
	req.Header.Set("Content-Type", "application/json")

	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	models.DB.First(&emailVerification, "email = ?", validEmail)

	assert.Equal(t, http.StatusCreated, rec.Code)
	assert.NotNil(t, emailVerification)
	assert.Equal(t, time.Now().Add(5*time.Minute).UTC().Truncate(time.Second), emailVerification.Expiration)
	assert.Contains(t, rec.Body.String(), "email was sent")

	rec = httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	assert.Equal(t, http.StatusBadRequest, rec.Code)
	assert.Contains(t, rec.Body.String(), "There is already a valid code for this email")
}

func TestVerifyEmailSuccess(t *testing.T) {
	SetupTestDB()

	emailDomain := factory.EmailDomainsFactory()
	validEmail := "email@" + emailDomain.Domain
	var emailVerification models.EmailsVerification
	models.DB.Create(&emailDomain)

	mockEmailSender := &mocks.MockSesSender{
		SendVerificationEmailFunc: func(recipient, subject, body string) error {
			return nil
		},
	}

	router := gin.Default()

	router.GET("/api/verify/email/:email", func(c *gin.Context) {
		controllers.VerifyEmail(c, mockEmailSender)
	})

	gin.SetMode(gin.TestMode)
	req, _ := http.NewRequest("GET", "/api/verify/email/"+validEmail, nil)
	req.Header.Set("Content-Type", "application/json")

	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	models.DB.First(&emailVerification, "email = ?", validEmail)

	assert.Equal(t, http.StatusCreated, rec.Code)
	assert.NotNil(t, emailVerification)
	assert.Equal(t, time.Now().Add(5*time.Minute).UTC().Truncate(time.Second), emailVerification.Expiration)
	assert.Contains(t, rec.Body.String(), "email was sent")
}
