package controllers

import (
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"unifriend-api/models"
	"unifriend-api/utils/aws"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type RegisterInput struct {
	Password          string `json:"password" binding:"required"`
	RePassword        string `json:"re_password" binding:"required"`
	Email             string `json:"email" binding:"required"`
	Name              string `json:"name" binding:"required"`
	ProfilePictureURL string `json:"profile_picture_url"`
	MajorID           uint   `json:"major_id" binding:"required"`
}

type ImageUploadInput struct {
	File   *multipart.FileHeader `form:"file" binding:"required"`
	UserId string                `form:"user_id" binding:"required"`
}

type LoginInput struct {
	Email    string `json:"email" binding:"required"`
	Password string `json:"password" binding:"required"`
}

type RegisterResponse struct {
	Message string `json:"message" example:"User created successfully"`
}

type LoginResponse struct {
	Token string `json:"token" example:"a34ojfds0cidsaokdjcdojfi"`
}

// @Description	Register
// @Accept			json
// @Tags			auth
// @Produce		json
// @Param			input	body		RegisterInput	true	"register input"
// @Success		201		{object}	controllers.RegisterResponse
// @Failure		400		"Invalid Data"
// @Router			/register [post]
func Register(c *gin.Context) {

	var input RegisterInput

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if models.UsernameAlreadyUsed(input.Email) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "something went wrong"})
		return
	}

	u := models.User{}

	if input.Password != input.RePassword {
		c.JSON(http.StatusBadRequest, gin.H{"error": "password and re_password are not the same"})
		return
	}

	u.Password = input.Password
	u.Email = input.Email
	u.Name = input.Name
	u.ProfilePictureURL = input.ProfilePictureURL
	u.MajorID = input.MajorID

	_, err := u.SaveUser()

	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, RegisterResponse{Message: "User created successfully"})
}

// @Description	Login
// @Accept			json
// @Tags			auth
// @Produce		json
// @Param			input	body		LoginInput	true	"login input"
// @Success		200		{object}	controllers.LoginResponse
// @Failure		400		"Invalid Data"
// @Failure		401		"email or password is incorrect.""
// @Router			/login [post]
func Login(c *gin.Context) {

	var input LoginInput

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	token, err := models.LoginCheck(input.Email, input.Password)

	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "username or password is incorrect."})
		return
	}

	c.JSON(http.StatusOK, LoginResponse{Token: token})

}

func UploadImage(c *gin.Context) {
	var imageValidator ImageUploadInput

	if err := c.ShouldBind(&imageValidator); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "file and user_id are required"})
		return
	}

	if !validateFileUploaded(imageValidator.File) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "could not upload the file"})
		return
	}

	UUId := uuid.New()
	fileName := UUId.String() + filepath.Ext(imageValidator.File.Filename)

	imageFile, _ := imageValidator.File.Open()
	uploadResult, err := aws.UploadFileToS3(imageFile, fileName)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save image"})
		return
	}

	userID, _ := strconv.ParseUint(imageValidator.UserId, 10, 32)
	user, userErr := models.GetUserByID(uint(userID))

	if userErr != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "user not found"})
		return
	}

	user.ProfilePictureURL = uploadResult
	_, saveUserErr := user.SaveUser()

	if saveUserErr != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to save user info"})
		aws.DeleteFileFromS3(fileName)
		return
	}

	c.JSON(http.StatusCreated, gin.H{"message": "image uploaded succesufuly"})
}

func validateFileUploaded(header *multipart.FileHeader) bool {
	fileExtension := verifyFileExtension(header)
	fileSize := verifyFileSize(header)
	return fileExtension && fileSize
}

func verifyFileExtension(header *multipart.FileHeader) bool {
	fileExtention := filepath.Ext(header.Filename)
	imageExtensions := []string{".jpeg", ".png", ".jpg"}

	for _, ext := range imageExtensions {
		if ext == fileExtention {
			return true
		}
	}

	return false
}

func verifyFileSize(header *multipart.FileHeader) bool {
	maxSize, _ := strconv.ParseInt(os.Getenv("MAX_SIZE_PROFILE_IMAGE_KB"), 10, 64)
	return (header.Size / 1000) <= maxSize
}