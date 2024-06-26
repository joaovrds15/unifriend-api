package controllers

import (
	"net/http"
	"unifriend-api/models"

	"github.com/gin-gonic/gin"
)

type RegisterInput struct {
	Password          string `json:"password" binding:"required"`
	RePassword        string `json:"re_password" binding:"required"`
	Email             string `json:"email" binding:"required"`
	Name              string `json:"name" binding:"required"`
	ProfilePictureURL string `json:"profile_picture_url"`
	MajorID           uint   `json:"major_id" binding:"required"`
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
