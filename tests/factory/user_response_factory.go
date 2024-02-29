package factory

import (
	"unifriend-api/models"
)

var questionResponseIDCounter int

func UserResponseFactory() models.UserResponse {
	questionResponseIDCounter++
	questionTable := models.UserResponse{
		QuestionID: QuestionTableFactory().ID,
		OptionID:   OptionTableFactory().ID,
		UserID:     UserFactory().ID,
	}
	questionTable.ID = uint(questionTableIDCounter)

	return questionTable
}
