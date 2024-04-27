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

type OptionsformatForResponse struct {
	Id   uint   `json:"id"`
	Text string `json:"option_text"`
}

type QuestionResponseFormat struct {
	Id      uint                       `json:"id"`
	Text    string                     `json:"text"`
	QuizId  uint                       `json:"quiz_id"`
	Options []OptionsformatForResponse `json:"options"`
}

func GetQuestions(c *gin.Context) {
	questions, err := models.GetQuestions()

	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	options, err := models.GetOptions()
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	optionsByQuestionId := make(map[uint][]OptionsformatForResponse)
	for _, option := range options {
		optionFormat := OptionsformatForResponse{
			Id:   option.ID,
			Text: option.Text,
		}

		optionsByQuestionId[option.QuestionID] = append(optionsByQuestionId[option.QuestionID], optionFormat)
	}

	var questionsResponse []QuestionResponseFormat
	for _, question := range questions {
		response := QuestionResponseFormat{
			Id:      question.ID,
			Text:    question.Text,
			QuizId:  question.Quiz_id,
			Options: optionsByQuestionId[question.ID],
		}
		questionsResponse = append(questionsResponse, response)
	}

	c.JSON(http.StatusOK, questionsResponse)
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
