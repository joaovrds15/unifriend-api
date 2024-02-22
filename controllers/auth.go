package controllers

import (
	"net/http"
	"unifriend-api/models"

	"github.com/gin-gonic/gin"
)

type RegisterInput struct {
	Username          string `json:"username" binding:"required"`
	Password          string `json:"password" binding:"required"`
	RePassword        string `json:"re_password" binding:"required"`
	Email             string `json:"email" binding:"required"`
	FirstName         string `json:"first_name" binding:"required"`
	LastName          string `json:"last_name" binding:"required"`
	ProfilePictureURL string `json:"profile_picture_url"`
	MajorID           uint   `json:"major_id" binding:"required"`
}

type LoginInput struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

func Register(c *gin.Context) {

	var input RegisterInput

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	u := models.User{}

	if input.Password != input.RePassword {
		c.JSON(http.StatusBadRequest, gin.H{"error": "password and re_password are not the same"})
		return
	}

	u.Username = input.Username
	u.Password = input.Password
	u.Email = input.Email
	u.FirstName = input.FirstName
	u.LastName = input.LastName
	u.ProfilePictureURL = input.ProfilePictureURL
	u.MajorID = input.MajorID

	_, err := u.SaveUser()

	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "registration success"})
}

func Login(c *gin.Context) {

	var input LoginInput

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	u := models.User{}

	u.Username = input.Username
	u.Password = input.Password

	token, err := models.LoginCheck(u.Username, u.Password)

	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "username or password is incorrect."})
		return
	}

	c.JSON(http.StatusOK, gin.H{"token": token})

}
