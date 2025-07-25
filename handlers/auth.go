package handlers

import (
	"net/http"
	"os"
	"strconv"
	"time"
	"unifriend-api/models"
	"unifriend-api/utils/token"

	"github.com/gin-gonic/gin"
)

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

type RegisterResponse struct {
	Message string `json:"message" example:"User created successfully"`
}

type LoginResponse struct {
	Token string `json:"token" example:"a34ojfds0cidsaokdjcdojfi"`
}

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

	user := models.User{}

	if input.Password != input.RePassword {
		c.JSON(http.StatusBadRequest, gin.H{"error": "password and re_password are not the same"})
		return
	}

	imagesUrl := []models.UsersImages{}

	user.Password = input.Password
	user.Email = input.Email
	user.Name = input.Name
	user.MajorID = input.MajorID
	user.PhoneNumber = input.PhoneNumber
	user.Images = imagesUrl

	_, err := user.SaveUser()

	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	models.DB.Preload("Major").First(&user, user.ID)

	token, err := token.GenerateToken(user.ID)

	if err != nil {
		c.JSON(http.StatusCreated, gin.H{"error" : "something went wrong"})
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

	response := UserLoginRegisterResponse{
		UserID:            user.ID,
		Name:              user.Name,
		Email:             user.Email,
		PhoneNumber:       user.PhoneNumber,
		ProfilePicture:    user.ProfilePictureURL,
		Major:             user.Major,
		Images:            user.Images,
	}

	c.JSON(http.StatusCreated, gin.H{"error" : false, "data" : response})
}

func Login(c *gin.Context) {

	var input LoginInput

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	token, user := models.LoginCheck(input.Email, input.Password)

	if user.ID == 0 {
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

	response := UserLoginRegisterResponse{
		UserID:            user.ID,
		Name:              user.Name,
		Email:             user.Email,
		PhoneNumber:       user.PhoneNumber,
		ProfilePicture:    user.ProfilePictureURL,
		Major:             user.Major,
		Images:            user.Images,
	}

	c.JSON(http.StatusOK, gin.H{"error" : false, "data" : response})
}

func Logout(c *gin.Context) {
	http.SetCookie(c.Writer, &http.Cookie{
		Name:     "auth_token",
		Value:    "",
		Path:     "/",
		Domain:   os.Getenv("CLIENT_DOMAIN"),
		Expires:  time.Unix(0, 0),
		Secure:   os.Getenv("GIN_MODE") == "release",
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
	})
	c.Status(http.StatusNoContent)
}