package factory

import (
	"unifriend-api/models"

	"github.com/go-faker/faker/v4"
)

func QuestionTableFactory() models.QuestionTable {
	questionTable := models.QuestionTable{
		Text:    faker.Word(),
		Quiz_id: QuizTableFactory().ID,
	}

	return questionTable
}
