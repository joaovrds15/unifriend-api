package handlers

import (
	"net/http"
	"unifriend-api/models"

	"github.com/gin-gonic/gin"
)

type GetMajorsReponse struct {
	Majors []models.Major `json:"majors"`
}

// PingExample godoc
//
//	@Description	get all majors registered
//	@Tags			major
//	@Accept			json
//	@Produce		json
//	@Success		200	{object}	controllers.GetMajorsReponse
//	@Router			/majors [get]
func GetMajors(c *gin.Context) {
	majors, err := models.GetMajors()

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "something went wrong"})
	}

	c.JSON(http.StatusOK, GetMajorsReponse{Majors: majors})
}
