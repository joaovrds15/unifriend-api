package tests

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"os"
	"strconv"
	"testing"
	"time"
	"unifriend-api/handlers"
	"unifriend-api/middleware"
	"unifriend-api/models"
	"unifriend-api/tests/factory"
	"unifriend-api/tests/mocks"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestDeleteUserImageSuccess(t *testing.T) {
	SetupTestDB()
	SetupRoutes()
	defer models.TearDownTestDB()

	user := factory.UserFactory()
	user.Images = nil
	models.DB.Create(&user)

	image := factory.UsersImagesFactory()
	image.User = user
	models.DB.Create(&image)

	mockUploader := &mocks.MockS3Uploader{
		DeleteImageFunc: func(fileName string) (error) {
			return nil
		},
	}
	router.Use(middleware.AuthMiddleware())
	router.DELETE("/api/users/images/:image_id", func(c *gin.Context) {
		c.Params = []gin.Param{{Key: "image_id", Value: strconv.Itoa(int(image.ID))}}
		handlers.DeleteUserImage(c, mockUploader)
	})

	req, _ := http.NewRequest("DELETE", "/api/users/images/1", nil) 

	authCookie := &http.Cookie{
		Name:  "auth_token",
		Value: factory.GetUserFactoryToken(user.ID),
		Path:  "/",
	}

	req.AddCookie(authCookie)

	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusNoContent, rec.Code)
	var deletedImage models.UsersImages
	err := models.DB.First(&deletedImage, image.ID).Error
	assert.Error(t, err)
	assert.Equal(t, "record not found", err.Error())
}

func TestDeleteUserImageMissingToken(t *testing.T) {
	SetupTestDB()
	SetupRoutes()
	defer models.TearDownTestDB()

	mockUploader := &mocks.MockS3Uploader{
		DeleteImageFunc: func(fileName string) (error) {
			return nil
		},
	}

	router.DELETE("/api/users/images/:image_id", func(c *gin.Context) {
		handlers.DeleteUserImage(c, mockUploader)
	})

	req, _ := http.NewRequest("DELETE", "/api/users/images/1", nil) 

	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusUnauthorized, rec.Code)
	assert.Contains(t, rec.Body.String(), "User ID not found in token")
}

func TestDeleteUserImageInvalidImageID(t *testing.T) {
	SetupTestDB()
	SetupRoutes()
	defer models.TearDownTestDB()

	mockUploader := &mocks.MockS3Uploader{
		DeleteImageFunc: func(fileName string) (error) {
			return nil
		},
	}

	user := factory.UserFactory()
	models.DB.Create(&user)

	router.Use(middleware.AuthMiddleware())
	router.DELETE("/api/users/images/:image_id", func(c *gin.Context) {
		handlers.DeleteUserImage(c, mockUploader)
	})

	req, _ := http.NewRequest("DELETE", "/api/users/images/-1", nil) 

	authCookie := &http.Cookie{
		Name:  "auth_token",
		Value: factory.GetUserFactoryToken(user.ID),
		Path:  "/",
	}

	req.AddCookie(authCookie)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusBadRequest, rec.Code)
	assert.Contains(t, rec.Body.String(), "Invalid image ID format")
}

func TestDeleteUserImageNotFound(t *testing.T) {
	SetupTestDB()
	defer models.TearDownTestDB()

	mockUploader := &mocks.MockS3Uploader{
		DeleteImageFunc: func(fileName string) (error) {
			return nil
		},
	}

	user := factory.UserFactory()
	models.DB.Create(&user)

	testRouter := gin.Default()
	testRouter.Use(func(c *gin.Context) { c.Set("user_id", user.ID); c.Next() })
	testRouter.DELETE("/api/users/me/images/:image_id", func(c *gin.Context) {
		handlers.DeleteUserImage(c, mockUploader)
	})

	req, _ := http.NewRequest("DELETE", "/api/users/me/images/99999", nil)

	rec := httptest.NewRecorder()
	testRouter.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusNotFound, rec.Code)
	assert.Contains(t, rec.Body.String(), "Image not found")
}

func TestDeleteUserImageAuthError(t *testing.T) {
	SetupTestDB()
	defer models.TearDownTestDB()

	mockUploader := &mocks.MockS3Uploader{
		DeleteImageFunc: func(fileName string) (error) {
			return nil
		},
	}

	user := factory.UserFactory()
	models.DB.Create(&user)

	otherUser := factory.UserFactory()
	models.DB.Create(&otherUser)

	userImage := factory.UsersImagesFactory()
	userImage.UserID = otherUser.ID
	models.DB.Create(&userImage)

	testRouter := gin.Default()
	testRouter.Use(func(c *gin.Context) { c.Set("user_id", user.ID); c.Next() })
	testRouter.DELETE("/api/users/me/images/:image_id", func(c *gin.Context) {
		handlers.DeleteUserImage(c, mockUploader)
	})

	req, _ := http.NewRequest("DELETE", "/api/users/me/images/"+strconv.Itoa(int(userImage.ID)), nil)

	rec := httptest.NewRecorder()
	testRouter.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusForbidden, rec.Code)
	assert.Contains(t, rec.Body.String(), "You are not authorized to delete this image")

	var notDeletedImage models.UsersImages
	err := models.DB.First(&notDeletedImage, userImage.ID).Error
	assert.Nil(t, err)
}

func TestAddUserImageS3Failure(t *testing.T) {
	SetupTestDB()
	defer models.TearDownTestDB()

	os.Setenv("MAX_SIZE_PROFILE_IMAGE_KB", "512")
	s3Error := "mock S3 error on add"
	mockUploader := &mocks.MockS3Uploader{
		UploadImageFunc: func(file multipart.File, fileName string) (string, error) {
			return "", errors.New(s3Error)
		},
	}

	user := factory.UserFactory()
	user.Images = nil
	models.DB.Create(&user)

	testRouter := gin.Default()
	testRouter.Use(func(c *gin.Context) { c.Set("user_id", user.ID); c.Next() })
	testRouter.POST("/api/users/me/images", func(c *gin.Context) {
		handlers.AddUserImage(c, mockUploader)
	})

	var body bytes.Buffer
	writer := multipart.NewWriter(&body)
	part, err := writer.CreateFormFile("file", "image.jpg")
	if err != nil {
		t.Fatalf("Failed to create form file: %v", err)
	}
	_, _ = part.Write([]byte("fake image data"))
	writer.Close()

	req, _ := http.NewRequest("POST", "/api/users/me/images", &body)
	req.Header.Set("Content-Type", writer.FormDataContentType())

	rec := httptest.NewRecorder()
	testRouter.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusInternalServerError, rec.Code)
	assert.Contains(t, rec.Body.String(), "Failed to upload image: "+s3Error)
	
	var count int64
	models.DB.Model(&models.UsersImages{}).Where("user_id = ?", user.ID).Count(&count)
	assert.Equal(t, int64(0), count)

	os.Unsetenv("MAX_SIZE_PROFILE_IMAGE_KB")
}

func TestAddUserImageInvalidFile(t *testing.T) {
	SetupTestDB()
	defer models.TearDownTestDB()
	os.Setenv("MAX_SIZE_PROFILE_IMAGE_KB", "1")

	mockUploader := &mocks.MockS3Uploader{}

	user := factory.UserFactory()
	models.DB.Create(&user)

	testRouter := gin.Default()
	testRouter.Use(func(c *gin.Context) { c.Set("user_id", user.ID); c.Next() })
	testRouter.POST("/api/users/me/images", func(c *gin.Context) {
		handlers.AddUserImage(c, mockUploader)
	})

	largeFile := bytes.Repeat([]byte("A"), 2*1024)
	var body bytes.Buffer
	writer := multipart.NewWriter(&body)
	part, err := writer.CreateFormFile("file", "large.jpg")
	if err != nil {
		t.Fatalf("Failed to create form file: %v", err)
	}
	_, _ = part.Write(largeFile)
	writer.Close()

	req, _ := http.NewRequest("POST", "/api/users/me/images", &body)
	req.Header.Set("Content-Type", writer.FormDataContentType())

	rec := httptest.NewRecorder()
	testRouter.ServeHTTP(rec, req)
	
	assert.Equal(t, http.StatusInternalServerError, rec.Code)
	assert.Contains(t, rec.Body.String(), "could not upload the file") 

	os.Unsetenv("MAX_SIZE_PROFILE_IMAGE_KB")
}

func TestAddUserImageSuccess(t *testing.T) {
	SetupTestDB()
	defer models.TearDownTestDB()

	os.Setenv("MAX_SIZE_PROFILE_IMAGE_KB", "512")
	expectedURL := "https://s3.com.br/new_gallery_image.jpg"

	mockUploader := &mocks.MockS3Uploader{
		UploadImageFunc: func(file multipart.File, fileName string) (string, error) {
			return expectedURL, nil
		},
	}

	user := factory.UserFactory()
	models.DB.Create(&user)

	testRouter := gin.Default()
	testRouter.Use(func(c *gin.Context) { c.Set("user_id", user.ID); c.Next() })
	testRouter.POST("/api/users/me/images", func(c *gin.Context) {
		handlers.AddUserImage(c, mockUploader)
	})

	var body bytes.Buffer
	writer := multipart.NewWriter(&body)
	part, err := writer.CreateFormFile("file", "gallery.jpg")
	if err != nil {
		t.Fatalf("Failed to create form file: %v", err)
	}
	_, _ = part.Write([]byte("fake image data"))
	writer.Close()

	req, _ := http.NewRequest("POST", "/api/users/me/images", &body)
	req.Header.Set("Content-Type", writer.FormDataContentType())

	rec := httptest.NewRecorder()
	testRouter.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusCreated, rec.Code)

	var responseImage models.UsersImages
	err = json.Unmarshal(rec.Body.Bytes(), &responseImage)
	assert.NoError(t, err)
	assert.Equal(t, expectedURL, responseImage.ImageUrl)
	assert.Equal(t, user.ID, responseImage.UserID)
	assert.NotZero(t, responseImage.ID)

	var dbImage models.UsersImages
	models.DB.First(&dbImage, responseImage.ID)
	assert.Equal(t, expectedURL, dbImage.ImageUrl)
	assert.Equal(t, user.ID, dbImage.UserID)

	os.Unsetenv("MAX_SIZE_PROFILE_IMAGE_KB")
}

func TestUpdateUserProfilePictureInvalidUserIDType(t *testing.T) {
	SetupTestDB()
	defer models.TearDownTestDB()

	mockUploader := &mocks.MockS3Uploader{}

	testRouter := gin.Default()
	testRouter.Use(func(c *gin.Context) { c.Set("user_id", "not-a-uint"); c.Next() })
	testRouter.PUT("/api/users/me/profile-picture", func(c *gin.Context) {
		handlers.UpdateUserProfilePicture(c, mockUploader)
	})

	req, _ := http.NewRequest("PUT", "/api/users/me/profile-picture", nil)

	rec := httptest.NewRecorder()
	testRouter.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusInternalServerError, rec.Code)
	assert.Contains(t, rec.Body.String(), "Invalid User ID format")
}

func TestUpdateUserProfilePictureInvalidFile(t *testing.T) {
	SetupTestDB()
	defer models.TearDownTestDB()
	os.Setenv("MAX_SIZE_PROFILE_IMAGE_KB", "1")

	mockUploader := &mocks.MockS3Uploader{
		DeleteImageFunc: func(fileName string) (error) {
			return nil
		},
	}

	user := factory.UserFactory()
	models.DB.Create(&user)

	testRouter := gin.Default()
	testRouter.Use(func(c *gin.Context) { c.Set("user_id", user.ID); c.Next() })
	testRouter.PUT("/api/users/me/profile-picture", func(c *gin.Context) {
		handlers.UpdateUserProfilePicture(c, mockUploader)
	})

	largeFile := bytes.Repeat([]byte("A"), 2*1024)
	var body bytes.Buffer
	writer := multipart.NewWriter(&body)
	part, err := writer.CreateFormFile("file", "test.jpg")
	if err != nil {
		t.Fatalf("Failed to create form file: %v", err)
	}
	_, _ = part.Write(largeFile)
	writer.Close()

	req, _ := http.NewRequest("PUT", "/api/users/me/profile-picture", &body)
	req.Header.Set("Content-Type", writer.FormDataContentType())

	rec := httptest.NewRecorder()
	testRouter.ServeHTTP(rec, req)
	assert.Equal(t, http.StatusInternalServerError, rec.Code) 
	assert.Contains(t, rec.Body.String(), "could not upload the file")

	os.Unsetenv("MAX_SIZE_PROFILE_IMAGE_KB")
}

func TestUpdateUserProfilePictureAuthError(t *testing.T) {
	SetupTestDB()
	defer models.TearDownTestDB()

	mockUploader := &mocks.MockS3Uploader{}

	router.Use(middleware.AuthMiddleware())
	router.PUT("/api/users/me/profile-picture", func(c *gin.Context) {
		handlers.UpdateUserProfilePicture(c, mockUploader)
	})

	var body bytes.Buffer
	writer := multipart.NewWriter(&body)
	part, err := writer.CreateFormFile("file", "profile.jpg")
	if err != nil {
		t.Fatalf("Failed to create form file: %v", err)
	}
	_, _ = part.Write([]byte("fake image data"))
	writer.Close()

	req, _ := http.NewRequest("PUT", "/api/users/me/profile-picture", &body)
	req.Header.Set("Content-Type", writer.FormDataContentType())

	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusUnauthorized, rec.Code)
	assert.Contains(t, rec.Body.String(), "Invalid Token")
}

func TestUpdateUserProfilePictureSuccess(t *testing.T) {
	SetupTestDB()
	os.Setenv("MAX_SIZE_PROFILE_IMAGE_KB", "512")
	defer models.TearDownTestDB()

	expectedURL := "https://s3.com.br/new_profile_pic.jpg"
	mockUploader := &mocks.MockS3Uploader{
		UploadImageFunc: func(file multipart.File, fileName string) (string, error) {
			return expectedURL, nil
		},
		DeleteImageFunc: func(fileName string) (error) {
			return nil
		},
	}

	user := factory.UserFactory()
	models.DB.Create(&user)

	var body bytes.Buffer
	writer := multipart.NewWriter(&body)
	part, err := writer.CreateFormFile("file", "profile.jpg")
	if err != nil {
		t.Fatalf("Failed to create form file: %v", err)
	}
	_, _ = part.Write([]byte("fake image data"))
	writer.Close()

	req, _ := http.NewRequest("PUT", "/users/profile-picture", &body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	authCookie := &http.Cookie{
		Name:  "auth_token",
		Value: factory.GetUserFactoryToken(user.ID),
	}

	req.AddCookie(authCookie)
	router.Use(middleware.AuthMiddleware())
	router.PUT("/users/profile-picture", func(c *gin.Context) {
		handlers.UpdateUserProfilePicture(c, mockUploader)
	})

	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)

	var response map[string]string
	err = json.Unmarshal(rec.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, expectedURL, response["profile_picture_url"])

	var updatedUser models.User
	models.DB.First(&updatedUser, user.ID)
	assert.Equal(t, expectedURL, updatedUser.ProfilePictureURL)
	os.Unsetenv("MAX_SIZE_PROFILE_IMAGE_KB")
}

func TestUploadImageS3Failure(t *testing.T) {
	SetupTestDB()
	defer models.TearDownTestDB()

	os.Setenv("MAX_SIZE_PROFILE_IMAGE_KB", "512")

	expectedError := "mock S3 upload error"
	mockUploader := &mocks.MockS3Uploader{
		UploadImageFunc: func(file multipart.File, fileName string) (string, error) {
			return "", errors.New(expectedError)
		},
	}

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	var body bytes.Buffer
	writer := multipart.NewWriter(&body)
	part, err := writer.CreateFormFile("file", "test.jpg")
	if err != nil {
		t.Fatalf("Failed to create form file: %v", err)
	}
	_, _ = part.Write([]byte("fake image data"))
	writer.Close()

	req, err := http.NewRequest("POST", "/", &body)
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}
	req.Header.Set("Content-Type", writer.FormDataContentType())
	c.Request = req

	url, uploadErr := handlers.UploadImage(c, mockUploader)

	assert.Empty(t, url)
	assert.NotNil(t, uploadErr)
	assert.Contains(t, uploadErr.Error(), expectedError)
	os.Unsetenv("MAX_SIZE_PROFILE_IMAGE_KB")
}

func TestUploadImageBiggerThenLimit(t *testing.T) {
	SetupTestDB()
	defer models.TearDownTestDB()

	mockUploader := &mocks.MockS3Uploader{
		UploadImageFunc: func(file multipart.File, fileName string) (string, error) {
			return "https:s3.com.br", nil
		},
	}

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	os.Setenv("MAX_SIZE_PROFILE_IMAGE_KB", "1")

	largeFile := bytes.Repeat([]byte("A"), 2*1024)
	var body bytes.Buffer
	writer := multipart.NewWriter(&body)
	part, err := writer.CreateFormFile("file", "test.jpg")
	if err != nil {
		t.Fatalf("Failed to create form file: %v", err)
	}
	_, _ = part.Write(largeFile)
	writer.Close()

	req, err := http.NewRequest("POST", "/", &body)
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}
	req.Header.Set("Content-Type", writer.FormDataContentType())
	c.Request = req

	url, uploadErr := handlers.UploadImage(c, mockUploader)

	assert.Empty(t, url)
	assert.NotNil(t, uploadErr)

	assert.Error(t, uploadErr)
	os.Unsetenv("MAX_SIZE_PROFILE_IMAGE_KB")
}

func TestUploadImageSuccess(t *testing.T) {
	SetupTestDB()
	defer models.TearDownTestDB()

	os.Setenv("MAX_SIZE_PROFILE_IMAGE_KB", "512")

	mockUploader := &mocks.MockS3Uploader{}

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	expectedURL := "https://s3.com.br/mockimage.jpg"
	mockUploader.UploadImageFunc = func(file multipart.File, fileName string) (string, error) {
		return expectedURL, nil
	}

	var body bytes.Buffer
	writer := multipart.NewWriter(&body)
	part, err := writer.CreateFormFile("file", "test.jpg")
	if err != nil {
		t.Fatalf("Failed to create form file: %v", err)
	}
	_, _ = part.Write([]byte("fake image data"))
	writer.Close()

	req, err := http.NewRequest("POST", "/", &body)
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}
	req.Header.Set("Content-Type", writer.FormDataContentType())
	c.Request = req

	url, uploadErr := handlers.UploadImage(c, mockUploader)

	assert.Equal(t, expectedURL, url)
	assert.Nil(t, uploadErr)
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

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	var body bytes.Buffer
	writer := multipart.NewWriter(&body)
	part, err := writer.CreateFormFile("file", "test.txt")
	if err != nil {
		t.Fatalf("Failed to create form file: %v", err)
	}
	_, _ = part.Write([]byte("fake image data"))
	writer.Close()

	req, err := http.NewRequest("POST", "/", &body)
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}
	req.Header.Set("Content-Type", writer.FormDataContentType())
	c.Request = req

	url, uploadErr := handlers.UploadImage(c, mockUploader)

	assert.Empty(t, url)
	assert.NotNil(t, uploadErr)
	assert.Error(t, uploadErr)
}

func TestUploadImageWithoutImage(t *testing.T) {
	SetupTestDB()
	defer models.TearDownTestDB()

	os.Setenv("MAX_SIZE_PROFILE_IMAGE_KB", "512")

	mockUploader := &mocks.MockS3Uploader{
		UploadImageFunc: func(file multipart.File, fileName string) (string, error) {
			return "Missing required key 'Body' in params", errors.New("Missing required key 'Body' in params")
		},
	}

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	var buffer bytes.Buffer
	writer := multipart.NewWriter(&buffer)
	writer.Close()

	req, err := http.NewRequest("POST", "/", &buffer)
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}
	req.Header.Set("Content-Type", writer.FormDataContentType())
	c.Request = req

	url, uploadErr := handlers.UploadImage(c, mockUploader)

	assert.Empty(t, url)
	assert.NotNil(t, uploadErr)
	assert.Contains(t, uploadErr.Error(), "could not upload the file")
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
		handlers.VerifyEmail(c, mockEmailSender)
	})

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
		handlers.VerifyEmail(c, mockEmailSender)
	})

	req, _ := http.NewRequest("GET", "/api/verify/email/invalidemail.com", nil)
	req.Header.Set("Content-Type", "application/json")

	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	assert.Equal(t, http.StatusBadRequest, rec.Code)
	assert.Contains(t, rec.Body.String(), "Email inválido")
}

func TestVerifyEmailAlreadyInUse(t *testing.T) {
	SetupTestDB()
	defer models.TearDownTestDB()

	user := factory.UserFactory()
	user.Email = "testuser@mail.com"

	models.DB.Create(&user)

	mockEmailSender := &mocks.MockSesSender{
		SendVerificationEmailFunc: func(recipient, subject, body string) error {
			return nil
		},
	}

	router := gin.Default()

	router.GET("/api/verify/email/:email", func(c *gin.Context) {
		handlers.VerifyEmail(c, mockEmailSender)
	})

	req, _ := http.NewRequest("GET", "/api/verify/email/"+user.Email, nil)
	req.Header.Set("Content-Type", "application/json")

	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	assert.Equal(t, http.StatusBadRequest, rec.Code)
	assert.Contains(t, rec.Body.String(), "Email inválido")
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
		handlers.VerifyEmail(c, mockEmailSender)
	})

	req, _ := http.NewRequest("GET", "/api/verify/email/"+invalidEmail, nil)
	req.Header.Set("Content-Type", "application/json")

	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	assert.Equal(t, http.StatusBadRequest, rec.Code)
	assert.Contains(t, rec.Body.String(), "Email inválido")
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
		handlers.VerifyEmail(c, mockEmailSender)
	})

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
		handlers.VerifyEmail(c, mockEmailSender)
	})

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
		handlers.VerifyEmail(c, mockEmailSender)
	})

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

	assert.Equal(t, http.StatusUnauthorized, rec.Code)
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

	assert.Equal(t, http.StatusUnauthorized, rec.Code)
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

func TestGetUserResponsesInvalidToken(t *testing.T) {

	SetupTestDB()
	defer models.TearDownTestDB()

	user := factory.UserFactory()
	userInvalid := factory.UserFactory()

	models.DB.Create(&user)
	models.DB.Create(&userInvalid)

	req, _ := http.NewRequest("GET", "/api/get-results/user/"+strconv.FormatUint(uint64(user.ID), 10), nil)
	req.Header.Set("Content-Type", "application/json")
	authCookie := &http.Cookie{
		Name:  "auth_token",
		Value: factory.GetUserFactoryToken(userInvalid.ID),
		Path:  "/",
	}

	req.AddCookie(authCookie)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusForbidden, rec.Code)
	expectedResponse := gin.H{"error": "You are not authorized to access this resource"}
	var actualResponse gin.H
	err := json.Unmarshal(rec.Body.Bytes(), &actualResponse)
	assert.NoError(t, err)
	assert.Equal(t, expectedResponse["error"], actualResponse["error"])
}

func TestGetUserResultEmptyResults(t *testing.T) {

	SetupTestDB()
	defer models.TearDownTestDB()

	user := factory.UserFactory()
	userResponses := factory.UserResponseFactory()
	models.DB.Create(&userResponses)
	models.DB.Create(&user)

	req, _ := http.NewRequest("GET", "/api/get-results/user/"+strconv.FormatUint(uint64(user.ID), 10), nil)
	req.Header.Set("Content-Type", "application/json")
	authCookie := &http.Cookie{
		Name:  "auth_token",
		Value: factory.GetUserFactoryToken(user.ID),
		Path:  "/",
	}

	req.AddCookie(authCookie)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	expectedResponse := gin.H{"data": make([]handlers.User, 0), "page" : 1, "limit" : 10, "total" : 0, "total_pages" : 0}
	fmt.Println(rec.Body.String())
	var actualResponse gin.H
	err := json.Unmarshal(rec.Body.Bytes(), &actualResponse)
	assert.NoError(t, err)
	assert.Equal(t, expectedResponse["error"], actualResponse["error"])
}

func TestGetUserResultWithMatches(t *testing.T) {
	SetupTestDB()
	defer models.TearDownTestDB()

	user := factory.UserFactory()
	models.DB.Create(&user)

	userResponse := factory.UserResponseFactory()
	userResponse.User = user
	models.DB.Create(&userResponse)

	otherUser := factory.UserFactory()
	models.DB.Create(&otherUser)

	matchingResponse := factory.UserResponseFactory()
	matchingResponse.User = otherUser
	matchingResponse.Question = userResponse.Question
	matchingResponse.Option = userResponse.Option
	models.DB.Create(&matchingResponse)

	req, _ := http.NewRequest("GET", "/api/get-results/user/"+strconv.FormatUint(uint64(user.ID), 10), nil)
	req.Header.Set("Content-Type", "application/json")
	authCookie := &http.Cookie{
		Name:  "auth_token",
		Value: factory.GetUserFactoryToken(user.ID),
		Path:  "/",
	}

	req.AddCookie(authCookie)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)

	var response gin.H
	err := json.Unmarshal(rec.Body.Bytes(), &response)
	assert.NoError(t, err)

	dataInterface, ok := response["data"].([]interface{})
	assert.True(t, ok, "data should be a slice")
	assert.NotEmpty(t, dataInterface)

	firstUser := dataInterface[0].(map[string]interface{})
	assert.Equal(t, float64(otherUser.ID), firstUser["user_id"])
	assert.Equal(t, otherUser.Name, firstUser["name"])
	assert.Equal(t, otherUser.ProfilePictureURL, firstUser["profile_picture_url"])
	assert.Equal(t, float64(1), firstUser["score"])
	assert.Equal(t, false, firstUser["has_pending_connection_request"])
	assert.Equal(t, false, firstUser["has_connection"])
}

func TestGetUserResultWithMatchesUserHasPendingConnectionRequests(t *testing.T) {
	SetupTestDB()
	defer models.TearDownTestDB()

	user := factory.UserFactory()
	models.DB.Create(&user)

	otherUser := factory.UserFactory()
	models.DB.Create(&otherUser)

	connectionRequest := factory.ConnectionRequestFactory()
	connectionRequest.RequestingUser = otherUser
	connectionRequest.RequestedUser = user
	models.DB.Create(&connectionRequest)

	quiz := factory.QuizTableFactory()
	models.DB.Create(&quiz)

	question := factory.QuestionTableFactory()
	question.Quiz = quiz
	models.DB.Create(&question)

	option := factory.OptionTableFactory()
	option.QuestionTable = question
	models.DB.Create(&option)

	userResponse := factory.UserResponseFactory()
	userResponse.User = user
	userResponse.Question = question
	userResponse.Option = option
	models.DB.Create(&userResponse)

	matchingResponse := factory.UserResponseFactory()
	matchingResponse.User = otherUser
	matchingResponse.Question = question
	matchingResponse.Option = option
	models.DB.Create(&matchingResponse)

	req, _ := http.NewRequest("GET", "/api/get-results/user/"+strconv.FormatUint(uint64(user.ID), 10), nil)
	req.Header.Set("Content-Type", "application/json")
	authCookie := &http.Cookie{
		Name:  "auth_token",
		Value: factory.GetUserFactoryToken(user.ID),
		Path:  "/",
	}

	req.AddCookie(authCookie)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)

	var response gin.H
	err := json.Unmarshal(rec.Body.Bytes(), &response)
	assert.NoError(t, err)

	dataInterface, ok := response["data"].([]interface{})
	assert.True(t, ok, "data should be a slice")
	assert.NotEmpty(t, dataInterface)

	firstUser := dataInterface[0].(map[string]interface{})
	assert.Equal(t, float64(otherUser.ID), firstUser["user_id"])
	assert.Equal(t, otherUser.Name, firstUser["name"])
	assert.Equal(t, otherUser.ProfilePictureURL, firstUser["profile_picture_url"])
	assert.Equal(t, float64(1), firstUser["score"])
	assert.Equal(t, true, firstUser["has_pending_connection_request"])
	assert.Equal(t, false, firstUser["has_connection"])
}

func TestGetUserResultsWithPagination(t *testing.T) {
    SetupTestDB()
    defer models.TearDownTestDB()

    user := factory.UserFactory()
    models.DB.Create(&user)

    userResponse := factory.UserResponseFactory()
    userResponse.User = user
    models.DB.Create(&userResponse)

    var otherUsers []models.User
    for i := 0; i < 15; i++ {
        otherUser := factory.UserFactory()
        models.DB.Create(&otherUser)
        otherUsers = append(otherUsers, otherUser)

        matchingResponse := factory.UserResponseFactory()
        matchingResponse.User = otherUser
        matchingResponse.Question = userResponse.Question
        matchingResponse.Option = userResponse.Option
        models.DB.Create(&matchingResponse)
    }

    // Test first page
    req, _ := http.NewRequest("GET", "/api/get-results/user/"+strconv.FormatUint(uint64(user.ID), 10)+"?page=1&limit=10", nil)
    req.Header.Set("Content-Type", "application/json")
    authCookie := &http.Cookie{
        Name:  "auth_token",
        Value: factory.GetUserFactoryToken(user.ID),
        Path:  "/",
    }

    req.AddCookie(authCookie)
    rec := httptest.NewRecorder()
    router.ServeHTTP(rec, req)

    assert.Equal(t, http.StatusOK, rec.Code)

    var response handlers.PaginatedUserResponse
    err := json.Unmarshal(rec.Body.Bytes(), &response)
    assert.NoError(t, err)

    assert.Equal(t, 10, len(response.Data))
    assert.Equal(t, 1, response.Page)
    assert.Equal(t, 10, response.Limit)
    assert.Equal(t, 15, response.Total)
    assert.Equal(t, 2, response.TotalPages)

    // Test second page
    req2, _ := http.NewRequest("GET", "/api/get-results/user/"+strconv.FormatUint(uint64(user.ID), 10)+"?page=2&limit=10", nil)
    req2.Header.Set("Content-Type", "application/json")
    req2.AddCookie(authCookie)
    rec2 := httptest.NewRecorder()
    router.ServeHTTP(rec2, req2)

    assert.Equal(t, http.StatusOK, rec2.Code)

    var response2 handlers.PaginatedUserResponse
    err = json.Unmarshal(rec2.Body.Bytes(), &response2)
    assert.NoError(t, err)

    assert.Equal(t, 5, len(response2.Data))
    assert.Equal(t, 2, response2.Page)
    assert.Equal(t, 10, response2.Limit)
    assert.Equal(t, 15, response2.Total)
    assert.Equal(t, 2, response2.TotalPages)
}

func TestGetUserResultsDefaultPagination(t *testing.T) {
    SetupTestDB()
    defer models.TearDownTestDB()

    user := factory.UserFactory()
    models.DB.Create(&user)

    userResponse := factory.UserResponseFactory()
    userResponse.User = user
    models.DB.Create(&userResponse)

    for i := 0; i < 5; i++ {
        otherUser := factory.UserFactory()
        models.DB.Create(&otherUser)

        matchingResponse := factory.UserResponseFactory()
        matchingResponse.User = otherUser
        matchingResponse.Question = userResponse.Question
        matchingResponse.Option = userResponse.Option
        models.DB.Create(&matchingResponse)
    }

    // Test without pagination parameters (should use defaults)
    req, _ := http.NewRequest("GET", "/api/get-results/user/"+strconv.FormatUint(uint64(user.ID), 10), nil)
    req.Header.Set("Content-Type", "application/json")
    authCookie := &http.Cookie{
        Name:  "auth_token",
        Value: factory.GetUserFactoryToken(user.ID),
        Path:  "/",
    }

    req.AddCookie(authCookie)
    rec := httptest.NewRecorder()
    router.ServeHTTP(rec, req)

    assert.Equal(t, http.StatusOK, rec.Code)

    var response handlers.PaginatedUserResponse
    err := json.Unmarshal(rec.Body.Bytes(), &response)
    assert.NoError(t, err)

    assert.Equal(t, 5, len(response.Data))
    assert.Equal(t, 1, response.Page) // Default page
    assert.Equal(t, 10, response.Limit) // Default limit
    assert.Equal(t, 5, response.Total)
    assert.Equal(t, 1, response.TotalPages)
}

func TestGetUserResultsInvalidPaginationParams(t *testing.T) {
    SetupTestDB()
    defer models.TearDownTestDB()

    user := factory.UserFactory()
    models.DB.Create(&user)

    // Test with invalid page and limit (should use defaults)
    req, _ := http.NewRequest("GET", "/api/get-results/user/"+strconv.FormatUint(uint64(user.ID), 10)+"?page=0&limit=-5", nil)
    req.Header.Set("Content-Type", "application/json")
    authCookie := &http.Cookie{
        Name:  "auth_token",
        Value: factory.GetUserFactoryToken(user.ID),
        Path:  "/",
    }

    req.AddCookie(authCookie)
    rec := httptest.NewRecorder()
    router.ServeHTTP(rec, req)

    assert.Equal(t, http.StatusOK, rec.Code)

    var response handlers.PaginatedUserResponse
    err := json.Unmarshal(rec.Body.Bytes(), &response)
    assert.NoError(t, err)

    assert.Equal(t, 1, response.Page) // Default page
    assert.Equal(t, 10, response.Limit) // Default limit
}

func TestGetUserResultsLimitExceeded(t *testing.T) {
    SetupTestDB()
    defer models.TearDownTestDB()

    user := factory.UserFactory()
    models.DB.Create(&user)

    userResponse := factory.UserResponseFactory()
    userResponse.User = user
    models.DB.Create(&userResponse)

    // Create 10 other users with matching responses
    for i := 0; i < 10; i++ {
        otherUser := factory.UserFactory()
        models.DB.Create(&otherUser)

        matchingResponse := factory.UserResponseFactory()
        matchingResponse.User = otherUser
        matchingResponse.Question = userResponse.Question
        matchingResponse.Option = userResponse.Option
        models.DB.Create(&matchingResponse)
    }

    // Test with limit exceeding maximum (should be capped at 100)
    req, _ := http.NewRequest("GET", "/api/get-results/user/"+strconv.FormatUint(uint64(user.ID), 10)+"?page=1&limit=150", nil)
    req.Header.Set("Content-Type", "application/json")
    authCookie := &http.Cookie{
        Name:  "auth_token",
        Value: factory.GetUserFactoryToken(user.ID),
        Path:  "/",
    }

    req.AddCookie(authCookie)
    rec := httptest.NewRecorder()
    router.ServeHTTP(rec, req)

    assert.Equal(t, http.StatusOK, rec.Code)

    var response handlers.PaginatedUserResponse
    err := json.Unmarshal(rec.Body.Bytes(), &response)
    assert.NoError(t, err)

    assert.Equal(t, 100, response.Limit) // Capped at maximum
    assert.Equal(t, 10, len(response.Data))
}

func TestGetUserResultsPageBeyondTotal(t *testing.T) {
    SetupTestDB()
    defer models.TearDownTestDB()

    user := factory.UserFactory()
    models.DB.Create(&user)

    userResponse := factory.UserResponseFactory()
    userResponse.User = user
    models.DB.Create(&userResponse)

    // Create 5 other users with matching responses
    for i := 0; i < 5; i++ {
        otherUser := factory.UserFactory()
        models.DB.Create(&otherUser)

        matchingResponse := factory.UserResponseFactory()
        matchingResponse.User = otherUser
        matchingResponse.Question = userResponse.Question
        matchingResponse.Option = userResponse.Option
        models.DB.Create(&matchingResponse)
    }

    // Test requesting page beyond total pages
    req, _ := http.NewRequest("GET", "/api/get-results/user/"+strconv.FormatUint(uint64(user.ID), 10)+"?page=5&limit=10", nil)
    req.Header.Set("Content-Type", "application/json")
    authCookie := &http.Cookie{
        Name:  "auth_token",
        Value: factory.GetUserFactoryToken(user.ID),
        Path:  "/",
    }

    req.AddCookie(authCookie)
    rec := httptest.NewRecorder()
    router.ServeHTTP(rec, req)

    assert.Equal(t, http.StatusOK, rec.Code)

    var response handlers.PaginatedUserResponse
    err := json.Unmarshal(rec.Body.Bytes(), &response)
    assert.NoError(t, err)

    assert.Equal(t, 0, len(response.Data)) // Empty data for page beyond total
    assert.Equal(t, 5, response.Page)
    assert.Equal(t, 10, response.Limit)
    assert.Equal(t, 5, response.Total)
    assert.Equal(t, 1, response.TotalPages)
}

func TestGetUserResultsEmptyResultsWithPagination(t *testing.T) {
    SetupTestDB()
    defer models.TearDownTestDB()

    user := factory.UserFactory()
    userResponses := factory.UserResponseFactory()
    models.DB.Create(&userResponses)
    models.DB.Create(&user)

    req, _ := http.NewRequest("GET", "/api/get-results/user/"+strconv.FormatUint(uint64(user.ID), 10)+"?page=1&limit=5", nil)
    req.Header.Set("Content-Type", "application/json")
    authCookie := &http.Cookie{
        Name:  "auth_token",
        Value: factory.GetUserFactoryToken(user.ID),
        Path:  "/",
    }

    req.AddCookie(authCookie)
    rec := httptest.NewRecorder()
    router.ServeHTTP(rec, req)

    assert.Equal(t, http.StatusOK, rec.Code)

    var response handlers.PaginatedUserResponse
    err := json.Unmarshal(rec.Body.Bytes(), &response)
    assert.NoError(t, err)

    assert.Equal(t, 0, len(response.Data))
    assert.Equal(t, 1, response.Page)
    assert.Equal(t, 5, response.Limit)
    assert.Equal(t, 0, response.Total)
    assert.Equal(t, 0, response.TotalPages)
}

func TestGetUserResultsInvalidPaginationQuery(t *testing.T) {
    SetupTestDB()
    defer models.TearDownTestDB()

    user := factory.UserFactory()
    models.DB.Create(&user)

    // Test with non-numeric pagination parameters
    req, _ := http.NewRequest("GET", "/api/get-results/user/"+strconv.FormatUint(uint64(user.ID), 10)+"?page=abc&limit=xyz", nil)
    req.Header.Set("Content-Type", "application/json")
    authCookie := &http.Cookie{
        Name:  "auth_token",
        Value: factory.GetUserFactoryToken(user.ID),
        Path:  "/",
    }

    req.AddCookie(authCookie)
    rec := httptest.NewRecorder()
    router.ServeHTTP(rec, req)

    assert.Equal(t, http.StatusBadRequest, rec.Code)
    assert.Contains(t, rec.Body.String(), "Invalid pagination Data")
}

func TestDeleteUser(t *testing.T) {
	SetupTestDB()
	defer models.TearDownTestDB()

	user := factory.UserFactory()
	user.SaveUser()

	mockUploader := &mocks.MockS3Uploader{
		DeleteImageFunc: func(fileName string) (error) {
			return nil
		},
	}

	testRouter := gin.Default()
	testRouter.Use(middleware.AuthMiddleware())
	testRouter.DELETE("/api/users/delete", func(c *gin.Context) {
		handlers.DeleteUserAccount(c, mockUploader)
	})

	authCookie := &http.Cookie{
		Name:  "auth_token",
		Value: factory.GetUserFactoryToken(user.ID),
		Path:  "/",
	}

	req, _ := http.NewRequest("DELETE", "/api/users/delete", nil)
	req.Header.Set("Content-Type", "application/json")

	req.AddCookie(authCookie)
	rec := httptest.NewRecorder()

	testRouter.ServeHTTP(rec, req)
	assert.Equal(t, http.StatusNoContent, rec.Code)

	var userAfter models.User
	models.DB.Find(&userAfter).Where("id = ?", user.ID)


	assert.NotNil(t, userAfter.DeletedAt)
	assert.Equal(t, 0, userAfter.Status)

}