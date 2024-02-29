package factory

import (
	"unifriend-api/models"

	"github.com/go-faker/faker/v4"
)

var quizIDCounter int

func QuizTableFactory() models.QuizTable {
	quizIDCounter++
	quiz := models.QuizTable{
		Title:       faker.Word(),
		Description: faker.Sentence(),
	}
	quiz.ID = uint(quizIDCounter)

	return quiz
}
