package factory

import (
	"unifriend-api/models"
)

func ConnectionFactory() models.Connection {
	connection := models.Connection{
		UserA: UserFactory(),
		UserB: UserFactory(),
		ConnectionRequest: ConnectionRequestFactory(),
	}

	models.DB.Create(&connection)

	return connection
}
