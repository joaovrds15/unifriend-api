package factory

import (
	"unifriend-api/models"

	"github.com/go-faker/faker/v4"
)

var OptionTableCounter int

func OptionTableFactory() models.OptionTable {
	OptionTableCounter++
	questionTable := models.OptionTable{
		Text:       faker.Word(),
		QuestionID: QuestionTableFactory().ID,
	}
	questionTable.ID = uint(OptionTableCounter)

	return questionTable
}

func OptionTableFactories(n int) []models.OptionTable {
	optionTables := make([]models.OptionTable, n)
	for i := 0; i < n; i++ {
		optionTables[i] = OptionTableFactory()
	}
	return optionTables
}
