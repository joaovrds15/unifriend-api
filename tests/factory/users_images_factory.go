package factory

import (
	"unifriend-api/models"

	"github.com/go-faker/faker/v4"
)

func UsersImagesFactory() models.UsersImages {
	userImages := models.UsersImages{
		ImageUrl: faker.URL(),
	}

	models.DB.Create(&userImages)

	return userImages
}
