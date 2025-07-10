package factory

import (
	"unifriend-api/models"

	"github.com/go-faker/faker/v4"
)

func OptionTableFactory() models.OptionTable {
	questionTable := models.OptionTable{
		Text:       faker.Word(),
		QuestionTable: QuestionTableFactory(),
	}

	models.DB.Create(&questionTable)

	return questionTable
}

func OptionTableFactories(n int) []models.OptionTable {
	optionTables := make([]models.OptionTable, n)
	for i := 0; i < n; i++ {
		optionTables[i] = OptionTableFactory()
	}
	return optionTables
}
