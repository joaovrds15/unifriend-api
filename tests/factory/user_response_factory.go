package factory

import (
	"unifriend-api/models"
)

func UserResponseFactory() models.UserResponse {
	questionTable := models.UserResponse{
		QuestionID: QuestionTableFactory().ID,
		OptionID:   OptionTableFactory().ID,
		UserID:     UserFactory().ID,
	}

	return questionTable
}
