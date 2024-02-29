package factory

import (
	"unifriend-api/models"

	"github.com/go-faker/faker/v4"
)

var majorIDCounter int

func MajorFactory() models.Major {
	majorIDCounter++
	major := models.Major{
		Name: faker.Word(),
	}
	major.ID = uint(majorIDCounter)

	return major
}
