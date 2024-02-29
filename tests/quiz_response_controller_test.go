package tests

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"
	"unifriend-api/models"
	"unifriend-api/tests/factory"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

//test get questions

func TestSaveAnswers(t *testing.T) {
	router := gin.Default()
	models.SetupTestDB()
	setupRoutes(router)

	user := factory.UserFactory()
	quiz := factory.QuizTableFactory()

	models.DB.Create(&quiz)
	models.DB.Create(&user)
	questions := [3]models.QuestionTable{}
	for i := 0; i < 3; i++ {
		question := factory.QuestionTableFactory()
		question.Quiz_id = quiz.ID
		models.DB.Create(&question)
		questions[i] = question
	}

	for _, question := range questions {
		options := factory.OptionTableFactories(5)
		models.DB.Model(&question).Association("Options").Append(&options)
	}

	payload := []byte(`{
		"user_id" : "1",
		"quiz_id" : "1",
		"answers" : [
			  {
				  "questionID": "1",
				  "selectedOptionID": "1"
			  },
			  {
				  "questionID": "2",
				  "selectedOptionID": "1"
			  },
			  {
				  "questionID": "3",
				  "selectedOptionID": "1"
			  }
		  ]
	  }`)

	req, _ := http.NewRequest("POST", "/api/answer/save", bytes.NewBuffer(payload))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+factory.GetUserFactoryToken(user.ID))

	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusCreated, rec.Code)

	var userResponses []models.UserResponse
	models.DB.Find(&userResponses).Where("user_id = ?", user.ID)
	for i, userResponse := range userResponses {
		assert.Equal(t, questions[i].ID, userResponse.QuestionID)
		assert.Equal(t, 1, userResponse.OptionID)
	}

	models.TearDownTestDB()
}
