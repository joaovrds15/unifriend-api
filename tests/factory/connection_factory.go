package factory

import (
	"unifriend-api/models"
)

func ConnectionFactory() models.Connection {
	connection := models.Connection{
		UserAID: UserFactory().ID,
		UserBID: UserFactory().ID,
	}

	return connection
}
