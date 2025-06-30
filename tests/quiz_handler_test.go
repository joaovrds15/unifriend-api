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

	"github.com/stretchr/testify/assert"
	"github.com/xeipuuv/gojsonschema"
)

func TestGetQuestions(t *testing.T) {
	SetupTestDB()
	defer models.TearDownTestDB()

	quiz := factory.QuizTableFactory()
	user := factory.UserFactory()
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
		models.DB.Create(&options)
		models.DB.Model(&question).Association("Options").Append(&options)
		question.Options = options
	}

	req, _ := http.NewRequest("GET", "/api/questions", nil)
	req.Header.Set("Content-Type", "application/json")

	authCookie := &http.Cookie{
		Name:  "auth_token",
		Value: factory.GetUserFactoryToken(user.ID),
		Path:  "/",
	}

	req.AddCookie(authCookie)

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
}

func TestGetQuestionsWhenUserAlreadyTakenQuiz(t *testing.T) {
	SetupTestDB()
	defer models.TearDownTestDB()

	quiz := factory.QuizTableFactory()
	user := factory.UserFactory()
	models.DB.Create(&quiz)
	models.DB.Create(&user)

	question := factory.QuestionTableFactory()

	options := factory.OptionTableFactories(2)
	models.DB.Create(&options)
	models.DB.Model(&question).Association("Options").Append(&options)
	question.Options = options

	userResponse := factory.UserResponseFactory()
	userResponse.OptionID = 1
	userResponse.QuestionID = question.ID
	userResponse.UserID = user.ID

	models.DB.Create(&userResponse)

	req, _ := http.NewRequest("GET", "/api/questions", nil)
	req.Header.Set("Content-Type", "application/json")

	authCookie := &http.Cookie{
		Name:  "auth_token",
		Value: factory.GetUserFactoryToken(user.ID),
		Path:  "/",
	}

	req.AddCookie(authCookie)

	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusForbidden, rec.Code)
}

func TestSaveAnswers(t *testing.T) {
	SetupTestDB()
	defer models.TearDownTestDB()

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
				  "question_id": 1,
				  "option_id": 1
			  },
			  {
				  "question_id": 2,
				  "option_id": 6
			  },
			  {
				  "question_id": 3,
				  "option_id": 11
			  }
		  ]
	  }`)

	req, _ := http.NewRequest("POST", "/api/answer/save", bytes.NewBuffer(payload))
	req.Header.Set("Content-Type", "application/json")
	authCookie := &http.Cookie{
		Name:  "auth_token",
		Value: factory.GetUserFactoryToken(user.ID),
		Path:  "/",
	}

	req.AddCookie(authCookie)
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

}
