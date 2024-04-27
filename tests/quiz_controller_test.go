package tests

import (
	"bytes"
	"log"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"testing"
	"unifriend-api/models"
	"unifriend-api/tests/factory"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/xeipuuv/gojsonschema"
)

func TestGetQuestions(t *testing.T) {
	router := gin.Default()
	models.SetupTestDB()
	setupRoutes(router)

	quiz := factory.QuizTableFactory()
	models.DB.Create(&quiz)

	questions := [3]models.QuestionTable{}
	for i := 0; i < 3; i++ {
		question := factory.QuestionTableFactory()
		question.Quiz_id = quiz.ID
		models.DB.Create(&question)
		questions[i] = question
	}

	for _, question := range questions {
		options := factory.OptionTableFactories(5)
		models.DB.Create(&options)
		models.DB.Model(&question).Association("Options").Append(&options)
		question.Options = options
	}

	req, _ := http.NewRequest("GET", "/api/question", nil)
	req.Header.Set("Content-Type", "application/json")

	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	absPath, err := filepath.Abs("json-schemas/test_get_questions.json")
	if err != nil {
		log.Fatalf("Error getting absolute path: %v", err)
	}

	schemaLoader := gojsonschema.NewReferenceLoader("file://" + absPath)

	loader := gojsonschema.NewStringLoader(rec.Body.String())
	result, err := gojsonschema.Validate(schemaLoader, loader)
	if err != nil {
		log.Fatalf("Error validating schema: %v", err)
	}

	assert.Equal(t, true, result.Valid())
	models.TearDownTestDB()
}

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
		"user_id" : 1,
		"quiz_id" : 1,
		"answers" : [
			  {
				  "questionID": 1,
				  "selectedOptionID": 1
			  },
			  {
				  "questionID": 2,
				  "selectedOptionID": 6
			  },
			  {
				  "questionID": 3,
				  "selectedOptionID": 11
			  }
		  ]
	  }`)

	req, _ := http.NewRequest("POST", "/api/private/answer/save", bytes.NewBuffer(payload))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+factory.GetUserFactoryToken(user.ID))
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)
	assert.Equal(t, http.StatusCreated, rec.Code)

	var userResponses []models.UserResponse
	models.DB.Find(&userResponses).Where("user_id = ?", user.ID)

	var count uint = 1
	for i, userResponse := range userResponses {
		assert.Equal(t, questions[i].ID, userResponse.QuestionID)
		assert.Equal(t, count, userResponse.OptionID)
		count += 5
	}

	models.TearDownTestDB()
}
