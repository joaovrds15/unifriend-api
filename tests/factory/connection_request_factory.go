package factory

import (
	"unifriend-api/models"
)

func ConnectionRequestFactory() models.ConnectionRequest {
	connectionRequest := models.ConnectionRequest{
		RequestingUserID: UserFactory().ID,
		RequestedUserID: UserFactory().ID,
	}

	return connectionRequest
}
