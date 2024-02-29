package factory

import (
	"unifriend-api/models"

	"github.com/go-faker/faker/v4"
)

var questionTableIDCounter int

func QuestionTableFactory() models.QuestionTable {
	questionTableIDCounter++
	questionTable := models.QuestionTable{
		Text:    faker.Word(),
		Quiz_id: QuizTableFactory().ID,
	}
	questionTable.ID = uint(questionTableIDCounter)

	return questionTable
}
