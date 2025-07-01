package handlers

import (
	"net/http"
	"unifriend-api/models"

	"github.com/gin-gonic/gin"
)

type SaveAnswersInput struct {
	QuizID  uint     `json:"quiz_id" binding:"required"`
	Answers []models.OptionTable `json:"answers" binding:"required"`
}

type Answer struct {
	QuestionID       uint `json:"question_id" binding:"required"`
	SelectedOptionID uint `json:"option_id" binding:"required"`
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

	hasTaken, err := models.HasUserAlreadyTakenQuiz(userIDUint)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if hasTaken {
		c.JSON(http.StatusForbidden, gin.H{"error": "You have already taken the quiz"})
		return
	}

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

	c.JSON(http.StatusOK, gin.H{ "error" : false, "data" : questionsResponse})
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
	userID, exists := c.Get("user_id")

	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User ID not found in token"})
		return
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	questionsWithOptions, err := models.GetQuestionsAndOptions(input.Answers)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if len(input.Answers) != len(questionsWithOptions) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "number of answers does not match number of questions"})
		return
	}

	userResponses := []models.UserResponse{}

	for _, answer := range input.Answers {

		userResponse := models.UserResponse{
			UserID:     userID.(uint),
			QuestionID: answer.QuestionID,
			OptionID:   answer.ID,
		}

		userResponses = append(userResponses, userResponse)
	}

	if err := models.SaveUserResponses(userResponses); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to save user responses"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"error":   false,
		"message": "answers saved successfully",
	})
}
