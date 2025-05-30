package controllers

import (
	"errors"
	"fmt"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"
	"unicode"
	"unifriend-api/models"
	"unifriend-api/services"
	"unifriend-api/utils/token"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"golang.org/x/exp/rand"
)

const Subject = "Unifriends Email de verificação"
const Message = "Seu código de verificação é: "

type ImageUploadInput struct {
	File *multipart.FileHeader `form:"file" binding:"required"`
}

type EmailCodeVerificationInput struct {
	Email            string `json:"email" binding:"required"`
	VerificationCode int    `json:"verification_code" binding:"required"`
}

type User struct {
	UserId            uint   `json:"user_id"`
	ProfilePictureURL string `json:"profile_picture_url"`
	Name              string `json:"name"`
	Score             int    `json:"score"`
}

type UserResponse struct {
	Data []User `json:"data"`
}

type RegisterInput struct {
	Password          string   `json:"password" binding:"required"`
	RePassword        string   `json:"re_password" binding:"required"`
	Email             string   `json:"email" binding:"required"`
	Name              string   `json:"name" binding:"required"`
	PhoneNumber       string   `json:"phone_number" binding:"required"`
	MajorID           uint     `json:"major_id" binding:"required"`
}

type LoginInput struct {
	Email    string `json:"email" binding:"required"`
	Password string `json:"password" binding:"required"`
}

type GetResultsInput struct {
	UserId uint `uri:"user_id" binding:"required"`
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

	imagesUrl := []models.UsersImages{}

	u.Password = input.Password
	u.Email = input.Email
	u.Name = input.Name
	u.MajorID = input.MajorID
	u.PhoneNumber = input.PhoneNumber
	u.Images = imagesUrl

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

func UploadImage(c *gin.Context, uploader services.S3Uploader) (string, error) {
	var imageValidator ImageUploadInput

	if err := c.ShouldBind(&imageValidator); err != nil {
		return "", errors.New("could not upload the file")
	}

	if !validateFileUploaded(imageValidator.File) {
		return "", errors.New("could not upload the file")
	}

	UUId := uuid.New()
	fileName := UUId.String() + filepath.Ext(imageValidator.File.Filename)

	imageFile, _ := imageValidator.File.Open()
	uploadResult, err := uploader.UploadImage(imageFile, fileName)

	if err != nil {
		return "", err
	}

	return uploadResult, nil
}

func UpdateUserProfilePicture(c *gin.Context, uploader services.S3Uploader) {
	userID, exists := c.Get("userID")

	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User ID not found in token"})
		return
	}

	userIDUint, ok := userID.(uint)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Invalid User ID format"})
		return
	}

	imageURL, err := UploadImage(c, uploader)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	var user models.User
	if err := models.DB.First(&user, userIDUint).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	user.ProfilePictureURL = imageURL
	if err := models.DB.Save(&user).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update user profile picture"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Profile picture updated successfully", "profile_picture_url": imageURL})
}

func AddUserImage(c *gin.Context, uploader services.S3Uploader) {
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User ID not found in token"})
		return
	}

	userIDUint, ok := userID.(uint)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Invalid User ID format"})
		return
	}

	imageURL, err := UploadImage(c, uploader)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to upload image: " + err.Error()})
		return
	}

	userImage := models.UsersImages{
		ImageUrl: imageURL,
		UserID:   userIDUint,
	}

	if err := models.DB.Create(&userImage).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save image record"})
		return
	}

	c.JSON(http.StatusCreated, userImage)
}

func DeleteUserImage(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User ID not found in token"})
		return
	}

	userIDUint, ok := userID.(uint)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Invalid User ID format"})
		return
	}

	imageIDStr := c.Param("image_id")
	imageID, err := strconv.ParseUint(imageIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid image ID format"})
		return
	}

	var userImage models.UsersImages
	if err := models.DB.First(&userImage, uint(imageID)).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Image not found"})
		return
	}

	if userImage.UserID != userIDUint {
		c.JSON(http.StatusForbidden, gin.H{"error": "You are not authorized to delete this image"})
		return
	}

	if err := models.DB.Delete(&userImage).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete image record"})
		return
	}

	c.Status(http.StatusNoContent)
}

func VerifyEmail(c *gin.Context, emailSender services.SesSender) {
	var emailVerificationInput VerifyEmailInput

	if err := c.BindUri(&emailVerificationInput); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "email parameter is required"})
		return
	}

	if models.UsernameAlreadyUsed(emailVerificationInput.Email) {
		c.JSON(http.StatusConflict, gin.H{"error": "Email já está em uso"})
		return
	}

	if !isValidEmail(emailVerificationInput.Email) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Email inválido"})
		return
	}

	if models.HasValidVerificationCode(emailVerificationInput.Email) {
		verificationCode, errCode := models.GetLastetVerificationCodeEmail(emailVerificationInput.Email)
		if errCode != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "verification code not found"})
			return
		}

		expirationTime := verificationCode.Expiration.Sub(time.Now().UTC()).Truncate(time.Second).Seconds()
		c.JSON(http.StatusOK, gin.H{"error": "There is already a valid code for this email", "expiration_time": expirationTime})
		return
	}

	rand.Seed(uint64(time.Now().UnixNano()))
	verificationCode := rand.Intn(900000) + 100000

	err := emailSender.SendVerificationEmail(emailVerificationInput.Email, Subject, Message+strconv.Itoa(verificationCode))

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "something went wrong"})
		return
	}

	verification, _ := models.SaveVerificationCode(emailVerificationInput.Email, verificationCode)
	expirationTime := verification.Expiration.Sub(time.Now().UTC()).Truncate(time.Second).Seconds()
	c.JSON(http.StatusCreated, gin.H{"message": "email was sent", "expiration_time": expirationTime})
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
		c.JSON(http.StatusUnauthorized, gin.H{"error": "code expired or is incorrect"})
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

func GetVerificationCodeExpiration(c *gin.Context) {
	var emailVerificationInput VerifyEmailInput

	if err := c.BindUri(&emailVerificationInput); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "email parameter is required"})
		return
	}

	verificationCode, errCode := models.GetLastetVerificationCodeEmail(emailVerificationInput.Email)

	if errCode != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "verification code not found"})
		return
	}

	expirationTime := verificationCode.Expiration.Sub(time.Now().UTC()).Truncate(time.Second).Seconds()
	if expirationTime > 0 {
		c.JSON(http.StatusOK, gin.H{"expiration_time": expirationTime})
		return
	}

	c.JSON(http.StatusOK, gin.H{"expiration_time": 0})
}

func GetResults(c *gin.Context) {
	var userGetResultsInput GetResultsInput
	if err := c.BindUri(&userGetResultsInput); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user_id parameter"})
		return
	}

	if err := validateUserAccess(c, userGetResultsInput.UserId); err != nil {
		c.JSON(http.StatusForbidden, gin.H{"error": "You are not authorized to access this resource"})
		return
	}

	currentUserResponses, err := models.GetUserResponsesByUserID(userGetResultsInput.UserId)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve your responses. Please try again."})
		return
	}

	if len(currentUserResponses) == 0 {
		c.JSON(200, gin.H{
			"error": false,
			"data":  make([]User, 0),
		})
		return
	}

	allMatchingResponsesFromOthers, err := models.GetMatchingResponsesFromOtherUsers(userGetResultsInput.UserId, currentUserResponses)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to compare responses. Please try again."})
		return
	}

	users := buildUserScores(allMatchingResponsesFromOthers)

	c.JSON(200, gin.H{
		"error": false,
		"data":  users,
	})
}

func validateUserAccess(c *gin.Context, requestedUserID uint) error {
	tokenUserIDClaim, _ := c.Get("user_id")
	tokenUserIDFloat, _ := tokenUserIDClaim.(float64)

	if uint(tokenUserIDFloat) != requestedUserID {
		return fmt.Errorf("unauthorized access")
	}

	return nil
}

func buildUserScores(matchingResponses []models.UserResponse) []User {
	scoreResult := make(map[uint]User)

	for _, matchingResponse := range matchingResponses {
		otherUserID := matchingResponse.UserID

		if user, exists := scoreResult[otherUserID]; !exists {
			scoreResult[otherUserID] = User{
				UserId:            otherUserID,
				Name:              matchingResponse.User.Name,
				ProfilePictureURL: matchingResponse.User.ProfilePictureURL,
				Score:             1,
			}
		} else {
			user.Score++
			scoreResult[otherUserID] = user
		}
	}

	var users []User
	for _, user := range scoreResult {
		users = append(users, user)
	}

	sort.Slice(users, func(i, j int) bool {
		return users[i].Score > users[j].Score
	})

	return users
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

	if len(emailParts) != 2 {
		return false
	}

	return models.EmailDomainExists(emailParts[1])
}
