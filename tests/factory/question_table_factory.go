package factory

import (
	"unifriend-api/models"

	"github.com/go-faker/faker/v4"
)

func QuestionTableFactory() models.QuestionTable {
	questionTable := models.QuestionTable{
		Text:    faker.Word(),
		Quiz: QuizTableFactory(),
		Options: OptionTableFactories(3),
	}

	return questionTable
}
