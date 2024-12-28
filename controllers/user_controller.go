package controllers

import (
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
	"unicode"
	"unifriend-api/models"
	"unifriend-api/services"
	"unifriend-api/utils/aws"
	"unifriend-api/utils/token"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"golang.org/x/exp/rand"
)

const Subject = "Unifriends Email de verificação"
const Message = "Seu código de verificação é: "

type ImageUploadInput struct {
	File   *multipart.FileHeader `form:"file" binding:"required"`
	UserId string                `form:"user_id" binding:"required"`
}

type EmailCodeVerificationInput struct {
	Email            string `json:"email" binding:"required"`
	VerificationCode int    `json:"verification_code" binding:"required"`
}

type RegisterInput struct {
	Password          string `json:"password" binding:"required"`
	RePassword        string `json:"re_password" binding:"required"`
	Email             string `json:"email" binding:"required"`
	Name              string `json:"name" binding:"required"`
	PhoneNumber       string `json:"phone_number" binding:"required"`
	ProfilePictureURL string `json:"profile_picture_url"`
	MajorID           uint   `json:"major_id" binding:"required"`
}

type LoginInput struct {
	Email    string `json:"email" binding:"required"`
	Password string `json:"password" binding:"required"`
}

type VerifyEmailInput struct {
	Email string `uri:"email" binding:"required"`
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

	if !isValidPassword(input.Password) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Password must be at least 8 characters long, contain at least one uppercase letter, and one special symbol"})
		return
	}

	if models.UsernameAlreadyUsed(input.Email) || models.PhoneNumberAlreadyUsed(input.PhoneNumber) {
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
	u.PhoneNumber = input.PhoneNumber

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

	tokenLifespanStr := os.Getenv("TOKEN_HOUR_LIFESPAN")
	tokenLifespan, _ := strconv.Atoi(tokenLifespanStr)

	c.SetCookie(
		"auth_token",
		token,
		tokenLifespan*3600,
		"/",
		os.Getenv("CLIENT_DOMAIN"),
		os.Getenv("GIN_MODE") == "release",
		true,
	)

	c.Status(http.StatusNoContent)
}

func UploadProfileImage(c *gin.Context, uploader services.S3Uploader) {
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
	uploadResult, err := uploader.UploadImage(imageFile, fileName)

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

func VerifyEmail(c *gin.Context, emailSender services.SesSender) {
	var emailVerificationInput VerifyEmailInput

	if err := c.BindUri(&emailVerificationInput); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "email parameter is required"})
		return
	}

	if !isValidEmail(emailVerificationInput.Email) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid email"})
		return
	}

	if models.HasValidExpirationCode(emailVerificationInput.Email) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "There is already a valid code for this email"})
		return
	}

	rand.Seed(uint64(time.Now().UnixNano()))
	verificationCode := rand.Intn(900000) + 100000

	err := emailSender.SendVerificationEmail(emailVerificationInput.Email, Subject, Message+strconv.Itoa(verificationCode))

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "something went wrong"})
		return
	}

	models.SaveVerificationCode(emailVerificationInput.Email, verificationCode)

	c.JSON(http.StatusCreated, gin.H{"message": "email was sent"})
}

func VerifyEmailCode(c *gin.Context) {
	var codeValidation EmailCodeVerificationInput

	if err := c.ShouldBind(&codeValidation); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "email or code is necessary"})
		return
	}

	verificationCode, errCode := models.GetLastetVerificationCodeEmail(codeValidation.Email)

	if errCode != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "verification code not found"})
		return
	}

	if !verificationCode.Expiration.After(time.Now().UTC()) || verificationCode.VerificationCode != codeValidation.VerificationCode {
		c.JSON(http.StatusBadRequest, gin.H{"error": "code expired or is incorrect"})
		return
	}

	registration := token.RegistrationToken{Email: codeValidation.Email}
	token, _ := registration.GenerateToken()

	registrationTokenLifespan, _ := strconv.Atoi(os.Getenv("TOKEN_REGISTRATION_HOUR_LIFESPAN"))

	c.SetCookie(
		"registration_token",
		token,
		registrationTokenLifespan*3600,
		"/",
		os.Getenv("CLIENT_DOMAIN"),
		os.Getenv("GIN_MODE") == "release",
		true,
	)

	c.Status(http.StatusCreated)
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

func isValidPassword(password string) bool {
	var hasMinLen, hasUpper, hasSpecial, hasNumber bool
	if len(password) >= 8 {
		hasMinLen = true
	}

	for _, char := range password {
		switch {
		case unicode.IsUpper(char):
			hasUpper = true
		case unicode.IsPunct(char) || unicode.IsSymbol(char):
			hasSpecial = true
		case unicode.IsNumber(char):
			hasNumber = true
		}
	}
	return hasMinLen && hasUpper && hasSpecial && hasNumber
}

func isValidEmail(email string) bool {

	emailParts := strings.Split(email, "@")

	if models.UsernameAlreadyUsed(email) {
		return false
	}

	if len(emailParts) != 2 {
		return false
	}

	return models.EmailDomainExists(emailParts[1])
}
