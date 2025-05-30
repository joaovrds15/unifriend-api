package factory

import (
	"unifriend-api/models"

	"github.com/go-faker/faker/v4"
)

func MajorFactory() models.Major {
	major := models.Major{
		Name: faker.Sentence(),
	}

	return major
}
