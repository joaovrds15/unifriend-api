package tests

import (
	"bytes"
	"encoding/json"
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

	for i := 0; i < 3; i++ {
		question := factory.QuestionTableFactory()
		question.Quiz = quiz
		models.DB.Create(&question)
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
	models.DB.Create(&question)
	options := factory.OptionTableFactories(2)
	options[0].QuestionTable = question
	options[1].QuestionTable = question
	models.DB.Create(&options)

	for _, option := range options {
		userResponse := factory.UserResponseFactory()
		userResponse.Option = option
		userResponse.Question = question
		userResponse.User = user

		models.DB.Create(&userResponse)
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
		question.Quiz = quiz
		models.DB.Create(&question)
		questions[i] = question
	}

	var request = make(map[string]any)
	request["user_id"] = user.ID
	request["quiz_id"] = quiz.ID

	answers := make([]map[string]any, 0)
	for _, question := range questions {
		answer := map[string]any{
			"question_id": question.ID,
			"option_id":   question.Options[0].ID,
		}
		answers = append(answers, answer)
	}
	request["answers"] = answers

	
	payload, err := json.Marshal(request)
	if err != nil {
		t.Fatalf("Failed to marshal request: %v", err)
	}

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
		count += 3
	}

}
