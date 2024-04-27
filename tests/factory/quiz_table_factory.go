package factory

import (
	"unifriend-api/models"

	"github.com/go-faker/faker/v4"
)

func QuizTableFactory() models.QuizTable {
	quiz := models.QuizTable{
		Title:       faker.Word(),
		Description: faker.Sentence(),
	}

	return quiz
}
