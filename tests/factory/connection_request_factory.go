package factory

import (
	"unifriend-api/models"
)

func ConnectionRequestFactory() models.ConnectionRequest {
	connectionRequest := models.ConnectionRequest{
		RequestingUser: UserFactory(),
		RequestedUser: UserFactory(),
	}

	return connectionRequest
}
