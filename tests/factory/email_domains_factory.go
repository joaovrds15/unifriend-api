package factory

import (
	"unifriend-api/models"

	"github.com/go-faker/faker/v4"
)

func EmailDomainsFactory() models.EmailDomains {
	emailDomain := models.EmailDomains{
		Institution: faker.LastName(),
		Domain:      faker.DomainName(),
	}

	return emailDomain
}
