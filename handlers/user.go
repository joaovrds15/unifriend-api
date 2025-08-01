package handlers

import (
	"errors"
	"fmt"
	"mime/multipart"
	"net/http"
	"os"
	"path"
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
	HasPendingConnectionRequest bool `json:"has_pending_connection_request"`
	HasConnection bool `json:"has_connection"`
	ConnectionRequest models.ConnectionRequest `json:"connection_request"`
}

type UserResponse struct {
	Data []User `json:"data"`
}

type GetResultsInput struct {
	UserId uint `uri:"user_id" binding:"required"`
	Page int `form:"page"`
	Limit int `form:"limit"`
}

type PaginatedUserResponse struct {
	Data []User `json:"data"`
	Page int `json:"page"`
	Limit int `json:"limit"`
	Total int `json:"total"`
	TotalPages int `json:"total_pages"`
}

type UserLoginRegisterResponse struct {
	UserID uint `json:"id"`
	Name string `json:"name"`
	Email string `json:"email"`
	PhoneNumber string `json:"phone_number"`
	ProfilePicture string `json:"profile_picture_url"`
	Major models.Major `json:"major"`
	Images []models.UsersImages `json:"images"`

}

type VerifyEmailInput struct {
	Email string `uri:"email" binding:"required"`
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

func deleteImage(filename string, uploader services.S3Uploader) (error) {
	if filename == "" {
		return nil
	}

	err := uploader.DeleteImage(path.Base(filename))
	if err != nil {
		return nil
	}

	return nil
}

func UpdateUserProfilePicture(c *gin.Context, uploader services.S3Uploader) {
	userID, exists := c.Get("user_id")

	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User ID not found in token"})
		return
	}

	userIDUint, ok := userID.(uint)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Invalid User ID format"})
		return
	}

	var user models.User
	if err := models.DB.First(&user, userIDUint).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	if user.ProfilePictureURL != "" {
		if err := deleteImage(user.ProfilePictureURL, uploader); err != nil {
			fmt.Printf("Warning: Failed to delete old profile picture: %v\n", err)
		}
	}

	imageURL, err := UploadImage(c, uploader)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	user.ProfilePictureURL = imageURL
	if err := models.DB.Save(&user).Error; err != nil {
		_ = deleteImage(imageURL, uploader)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update user profile picture"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Profile picture updated successfully", "profile_picture_url": imageURL})
}

func AddUserImage(c *gin.Context, uploader services.S3Uploader) {
	userID, exists := c.Get("user_id")
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

func DeleteUserImage(c *gin.Context, uploader services.S3Uploader) {
	userIDUint, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User ID not found in token"})
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

	if err := deleteImage(userImage.ImageUrl, uploader); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete image from storage"})
		return
	}

	if err := models.DB.Delete(&userImage).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete image record"})
		return
	}

	c.Status(http.StatusNoContent)
}

func DeleteUserProfilePicture(c *gin.Context, uploader services.S3Uploader) {
	userIDUint, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User ID not found in token"})
		return
	}

	var user models.User
	if err := models.DB.First(&user, userIDUint).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Image not found"})
		return
	}

	userProfilePicture := user.ProfilePictureURL

	if user.ProfilePictureURL == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "You you have not profile picture"})
		return
	}

	user.ProfilePictureURL = ""

	if err := models.DB.Save(&user).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete image record"})
		return
	}

	if err := deleteImage(userProfilePicture, uploader); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete image from storage"})
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

	if !isValidEmail(emailVerificationInput.Email) || models.UsernameAlreadyUsed(emailVerificationInput.Email) {
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

	verificationCode.Verified = true
	models.DB.Save(&verificationCode)

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

	if verificationCode.Verified {
        c.JSON(http.StatusConflict, gin.H{"error": "email has already been verified"})
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

	if err := c.ShouldBindQuery(&userGetResultsInput); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error" : "Invalid pagination Data"})
		return
	}

	if userGetResultsInput.Page <= 0 {
        userGetResultsInput.Page = 1
    }
    if userGetResultsInput.Limit <= 0 {
        userGetResultsInput.Limit = 10
    }
    if userGetResultsInput.Limit > 100 {
        userGetResultsInput.Limit = 100 // Max limit
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
		c.JSON(200, PaginatedUserResponse{
			Data: make([]User, 0),
			Page: userGetResultsInput.Page,
			Limit: userGetResultsInput.Limit,
			Total: 0,
			TotalPages: 0,
		})
		return
	}

	allMatchingResponsesFromOthers, err := models.GetMatchingResponsesFromOtherUsers(userGetResultsInput.UserId, currentUserResponses)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to compare responses. Please try again."})
		return
	}

	allUsers := buildUserScores(allMatchingResponsesFromOthers)

	total := len(allUsers)
    totalPages := (total + userGetResultsInput.Limit - 1) / userGetResultsInput.Limit

	offset := (userGetResultsInput.Page - 1) * userGetResultsInput.Limit
    
    var paginatedUsers []User
    if offset < total {
        end := offset + userGetResultsInput.Limit
        if end > total {
            end = total
        }
        paginatedUsers = allUsers[offset:end]
    } else {
        paginatedUsers = make([]User, 0)
    }

    response := PaginatedUserResponse{
        Data:       paginatedUsers,
        Page:       userGetResultsInput.Page,
        Limit:      userGetResultsInput.Limit,
        Total:      total,
        TotalPages: totalPages,
    }

    c.JSON(200, response)
}

func validateUserAccess(c *gin.Context, requestedUserID uint) error {
	tokenUserIDClaim, _ := c.Get("user_id")
	tokenUserIDUint, _ := tokenUserIDClaim.(uint)

	if tokenUserIDUint != requestedUserID {
		return fmt.Errorf("unauthorized access")
	}

	return nil
}

func buildUserScores(matchingResponses []models.MatchingUserResponse) []User {
	scoreResult := make(map[uint]User)

	for _, matchingResponse := range matchingResponses {
		otherUserID := matchingResponse.UserID

		if user, exists := scoreResult[otherUserID]; !exists {
			scoreResult[otherUserID] = User{
				UserId:            otherUserID,
				Name:              matchingResponse.User.Name,
				ProfilePictureURL: matchingResponse.User.ProfilePictureURL,
				Score:             1,
				HasPendingConnectionRequest: matchingResponse.HasPendingConnectionRequest,
				HasConnection: matchingResponse.HasConnection,
				ConnectionRequest: matchingResponse.ConnectionRequest,
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

func DeleteUserAccount(c *gin.Context, uploader services.S3Uploader) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User ID not found in token"})
		return
	}

	userIDUint, ok := userID.(uint)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Invalid User ID format"})
		return
	}

	var user models.User
	if err := models.DB.First(&user, userIDUint).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	if user.ProfilePictureURL != "" {
		if err := deleteImage(user.ProfilePictureURL, uploader); err != nil {
			fmt.Printf("Warning: Failed to delete profile picture: %v\n", err)
		}
	}

	if err := user.DeleteUser(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete user account"})
		return
	}

	c.Status(http.StatusNoContent)
}