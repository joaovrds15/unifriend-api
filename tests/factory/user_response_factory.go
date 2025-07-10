package factory

import (
	"unifriend-api/models"
)

func UserResponseFactory() models.UserResponse {
	questionTable := models.UserResponse{
		Question: QuestionTableFactory(),
		Option:   OptionTableFactory(),
		User:     UserFactory(),
	}

	return questionTable
}
