package controllers

import (
	"net/http"
	"unifriend-api/models"

	"github.com/gin-gonic/gin"
)

type SaveAnswersInput struct {
	QuizID  uint     `json:"quiz_id" binding:"required"`
	UserID  uint     `json:"user_id" binding:"required"`
	Answers []Answer `json:"answers" binding:"required"`
}

type Answer struct {
	QuestionID       uint `json:"questionID" binding:"required"`
	SelectedOptionID uint `json:"selectedOptionID" binding:"required"`
}

func SaveAnswers(c *gin.Context) {
	var input SaveAnswersInput

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	user, err := models.GetUserByID(input.UserID)

	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	for _, answer := range input.Answers {
		question, err := models.GetQuestionByID(answer.QuestionID)

		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		option, err := models.GetOptionByID(answer.SelectedOptionID)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		if option.QuestionID != question.ID {
			c.JSON(http.StatusBadRequest, gin.H{"error": "selected option does not belong to the question"})
			return
		}

		userResponse := models.UserResponse{
			UserID:     user.ID,
			QuestionID: question.ID,
			OptionID:   option.ID,
		}

		userResponse.SaveUserResponse()
	}

	c.JSON(http.StatusCreated, gin.H{"message": "answers saved successfully"})
}
