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

type SaveAnswerResponse struct {
	Message string `json:"message" example:"answers saved successfully"`
}
type OptionsformatForResponse struct {
	Id   uint   `json:"id" example:"1"`
	Text string `json:"option_text" example:"clubbing"`
}

type QuestionResponseFormat struct {
	Id      uint                       `json:"id" example:"1"`
	Text    string                     `json:"text" example:"Best Place to go out on weekends?"`
	QuizId  uint                       `json:"quiz_id" example:"1"`
	Options []OptionsformatForResponse `json:"options"`
}

// @Description	Get Quiz questions
// @Accept			json
// @Tags			quiz
// @Produce		json
// @Security		Bearer
// @Success		200	{object}	controllers.QuestionResponseFormat
// @Failure		500	"Something went wrong"
// @Router			/questions [get]
func GetQuestions(c *gin.Context) {
	questions, err := models.GetQuestions()

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	options, err := models.GetOptions()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
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

	if questionsResponse == nil {
		c.JSON(http.StatusOK, gin.H{})
		return
	}

	c.JSON(http.StatusOK, questionsResponse)
}

// @Description	SaveAnswers
// @Accept			json
// @Tags			quiz
// @Produce		json
// @Param			input	body		SaveAnswersInput	true	"Save answers input"
// @Success		201		{object}	controllers.SaveAnswerResponse
// @Failure		400		"Invalid Data"
// @Failure		500	"Something went wrong"
// @Security		Bearer
// @Router			/answer/save [post]
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
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		option, err := models.GetOptionByID(answer.SelectedOptionID)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		if option.QuestionID != question.ID {
			c.JSON(http.StatusBadRequest, gin.H{"error": "selected option does not belong to this question"})
			return
		}

		userResponse := models.UserResponse{
			UserID:     user.ID,
			QuestionID: question.ID,
			OptionID:   option.ID,
		}

		userResponse.SaveUserResponse()
	}

	c.JSON(http.StatusCreated, SaveAnswerResponse{})
}
