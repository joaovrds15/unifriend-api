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
	image.UserID = user.ID
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
		controllers.DeleteUserImage(c, mockUploader)
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
		controllers.DeleteUserImage(c, mockUploader)
	})

	req, _ := http.NewRequest("DELETE", "/api/users/images/1", nil) 

	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusUnauthorized, rec.Code)
	assert.Contains(t, rec.Body.String(), "User ID not found in token")
}

func TestDeleteUserImageInvalidImageID(t *testing.T) {
	SetupTestDB()
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
		controllers.DeleteUserImage(c, mockUploader)
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
		controllers.DeleteUserImage(c, mockUploader)
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
		controllers.DeleteUserImage(c, mockUploader)
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
		controllers.AddUserImage(c, mockUploader)
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
		controllers.AddUserImage(c, mockUploader)
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
		controllers.AddUserImage(c, mockUploader)
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
		controllers.UpdateUserProfilePicture(c, mockUploader)
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
		controllers.UpdateUserProfilePicture(c, mockUploader)
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
		controllers.UpdateUserProfilePicture(c, mockUploader)
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
		controllers.UpdateUserProfilePicture(c, mockUploader)
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

	url, uploadErr := controllers.UploadImage(c, mockUploader)

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

	url, uploadErr := controllers.UploadImage(c, mockUploader)

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

	url, uploadErr := controllers.UploadImage(c, mockUploader)

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

	url, uploadErr := controllers.UploadImage(c, mockUploader)

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

	url, uploadErr := controllers.UploadImage(c, mockUploader)

	assert.Empty(t, url)
	assert.NotNil(t, uploadErr)
	assert.Contains(t, uploadErr.Error(), "could not upload the file")
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

	userData := map[string]interface{}{
		"password":            "Senha@123",
		"re_password":         "Senha@123",
		"major_id":            1,
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

	major := factory.MajorFactory()
	user := factory.UserFactory()
	user.Email = "testuser@mail.com"

	models.DB.Create(&major)
	models.DB.Create(&user)

	userData := map[string]interface{}{
		"password":            "Senha@123",
		"re_password":         "Senha@123",
		"major_id":            1,
		"email":               "testuser@mail.com",
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

	imagesUrls := []string{
		"http://test.com",
		"http://test2.com",
	}

	userData := map[string]interface{}{
		"password":            "Senha123",
		"re_password":         "Senha123",
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
		controllers.VerifyEmail(c, mockEmailSender)
	})

	req, _ := http.NewRequest("GET", "/api/verify/email/"+user.Email, nil)
	req.Header.Set("Content-Type", "application/json")

	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	assert.Equal(t, http.StatusConflict, rec.Code)
	assert.Contains(t, rec.Body.String(), "Email já está em uso")
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
		controllers.VerifyEmail(c, mockEmailSender)
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
		controllers.VerifyEmail(c, mockEmailSender)
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
		controllers.VerifyEmail(c, mockEmailSender)
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
	expectedResponse := gin.H{"error": false, "data": make([]controllers.User, 0)}
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
	userResponse.UserID = user.ID
	models.DB.Create(&userResponse)

	otherUser := factory.UserFactory()
	models.DB.Create(&otherUser)

	matchingResponse := factory.UserResponseFactory()
	matchingResponse.UserID = otherUser.ID
	matchingResponse.QuestionID = userResponse.QuestionID
	matchingResponse.OptionID = userResponse.OptionID
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

	assert.Equal(t, false, response["error"])

	dataInterface, ok := response["data"].([]interface{})
	assert.True(t, ok, "data should be a slice")
	assert.NotEmpty(t, dataInterface)

	firstUser := dataInterface[0].(map[string]interface{})
	assert.Equal(t, float64(otherUser.ID), firstUser["user_id"])
	assert.Equal(t, otherUser.Name, firstUser["name"])
	assert.Equal(t, otherUser.ProfilePictureURL, firstUser["profile_picture_url"])
	assert.Equal(t, float64(1), firstUser["score"])
}
