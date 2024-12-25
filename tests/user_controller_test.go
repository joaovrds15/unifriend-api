package tests

import (
	"bytes"
	"encoding/json"
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
	defer models.TearDownTestDB()

	os.Setenv("MAX_SIZE_PROFILE_IMAGE_KB", "512")

	mockUploader := &mocks.MockS3Uploader{
		UploadImageFunc: func(file multipart.File, fileName string) (string, error) {
			return "Missing required key 'Body' in params", errors.New("Missing required key 'Body' in params")
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
	assert.Contains(t, rec.Body.String(), "file is required")
}

func TestUploadImageInvalidExtension(t *testing.T) {
	SetupTestDB()
	defer models.TearDownTestDB()

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
	defer models.TearDownTestDB()

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
	assert.Contains(t, rec.Body.String(), "https:s3.com.br")
}

func TestUploadImageBiggerThenLimit(t *testing.T) {
	SetupTestDB()
	defer models.TearDownTestDB()

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

	imagesUrls := []string{
		"http://test.com",
		"http://test2.com",
	}

	userData := map[string]interface{}{
		"password":            "Senha@123",
		"re_password":         "Senha@123",
		"major_id":            1,
		"email":               "testemail@mail.com",
		"name":                "test user",
		"profile_picture_url": "http://test.com",
		"phone_number":        "62999999999",
		"images":              imagesUrls,
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

	var userImage models.UsersImages
	models.DB.Where("user_id = ?", user.ID).First(&userImage)
	assert.NotNil(t, userImage)
	assert.Equal(t, user.ID, userImage.UserID)

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

	imagesUrls := []string{
		"http://test.com",
		"http://test2.com",
	}

	userData := map[string]interface{}{
		"password":            "Senha@123",
		"re_password":         "Senha@123",
		"major_id":            1,
		"email":               "testuser@mail.com",
		"name":                "test user",
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

func TestRegisterWithDuplicatedPhoneNumber(t *testing.T) {
	SetupTestDB()
	defer models.TearDownTestDB()

	major := factory.MajorFactory()
	user := factory.UserFactory()
	user.PhoneNumber = "62999999999"

	models.DB.Create(&major)
	models.DB.Create(&user)

	userData := map[string]interface{}{
		"password":            "Senha@123",
		"re_password":         "Senha@123",
		"major_id":            1,
		"email":               "newuser@mail.com",
		"name":                "new user",
		"profile_picture_url": "http://test.com",
		"phone_number":        "62999999999",
		"images":              []models.UsersImages{factory.UsersImagesFactory()},
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
		"major_id":            1,
		"email":               "testemail@mail.com",
		"name":                "test user",
		"profile_picture_url": "http://test.com",
		"phone_number":        "62999999999",
		"images":              []models.UsersImages{factory.UsersImagesFactory()},
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

func TestVerifyEmailWithoutEmailParameter(t *testing.T) {
	SetupTestDB()
	defer models.TearDownTestDB()

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
	defer models.TearDownTestDB()

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
	defer models.TearDownTestDB()

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
	defer models.TearDownTestDB()

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
	defer models.TearDownTestDB()

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
	defer models.TearDownTestDB()

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

func TestVerifyEmailWithoutEmail(t *testing.T) {
	SetupTestDB()
	defer models.TearDownTestDB()

	req, _ := http.NewRequest("POST", "/api/verify/email", nil)
	req.Header.Set("Content-Type", "application/json")

	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusBadRequest, rec.Code)
	assert.Contains(t, rec.Body.String(), "email or code is necessary")
}

func TestVerifyEmailCodeDIfferentThanCreated(t *testing.T) {
	SetupTestDB()

	defer models.TearDownTestDB()

	emailVerification := factory.EmailsVerificationFactory()

	models.DB.Create(&emailVerification)

	userData := map[string]interface{}{
		"email":             emailVerification.Email,
		"verification_code": -1,
	}

	jsonValue, _ := json.Marshal(userData)

	req, _ := http.NewRequest("POST", "/api/verify/email", bytes.NewBuffer(jsonValue))
	req.Header.Set("Content-Type", "application/json")

	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusBadRequest, rec.Code)
	assert.Contains(t, rec.Body.String(), "code expired or is incorrect")
}

func TestVerifyEmailExpiredCode(t *testing.T) {
	SetupTestDB()

	defer models.TearDownTestDB()

	os.Setenv("VERIFICATION_CODE_LIFESPAN_MINUTES", "1")

	emailVerification := factory.EmailsVerificationFactory()
	emailVerification.Expiration = time.Now().Add(-1 * time.Minute).UTC().Truncate(time.Second)
	models.DB.Create(&emailVerification)

	userData := map[string]interface{}{
		"email":             emailVerification.Email,
		"verification_code": emailVerification.VerificationCode,
	}

	jsonValue, _ := json.Marshal(userData)

	req, _ := http.NewRequest("POST", "/api/verify/email", bytes.NewBuffer(jsonValue))
	req.Header.Set("Content-Type", "application/json")

	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusBadRequest, rec.Code)
	assert.Contains(t, rec.Body.String(), "code expired or is incorrect")
}

func TestVerifyEmailCodeSuccess(t *testing.T) {
	SetupTestDB()

	defer models.TearDownTestDB()

	os.Setenv("VERIFICATION_CODE_LIFESPAN_MINUTES", "1")

	emailVerification := factory.EmailsVerificationFactory()
	emailVerification.Expiration = time.Now().Add(1 * time.Minute).UTC().Truncate(time.Second)
	models.DB.Create(&emailVerification)

	userData := map[string]interface{}{
		"email":             emailVerification.Email,
		"verification_code": emailVerification.VerificationCode,
	}

	jsonValue, _ := json.Marshal(userData)

	req, _ := http.NewRequest("POST", "/api/verify/email", bytes.NewBuffer(jsonValue))
	req.Header.Set("Content-Type", "application/json")

	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusCreated, rec.Code)
	assert.NotEmpty(t, rec.Header().Get("Set-Cookie"))
	assert.Contains(t, rec.Header().Get("Set-Cookie"), "registration_token")
}
