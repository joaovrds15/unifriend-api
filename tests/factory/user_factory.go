package factory

import (
	"os"
	"unifriend-api/models"
	"unifriend-api/utils/token"

	"github.com/go-faker/faker/v4"
)

var userIDCounter int

func UserFactory() models.User {
	userIDCounter++
	user := models.User{
		Username:          faker.Username(),
		Password:          faker.Password(),
		Email:             faker.Email(),
		FirstName:         faker.FirstName(),
		LastName:          faker.LastName(),
		ProfilePictureURL: faker.URL(),
		IsAdmin:           false,
		MajorID:           MajorFactory().ID,
	}
	user.ID = uint(userIDCounter)

	return user
}

func GetUserFactoryToken(user_id uint) string {
	os.Setenv("TOKEN_HOUR_LIFESPAN", "1")
	os.Setenv("API_SECRET", "secret")
	token, error := token.GenerateToken(uint(user_id))

	if error == nil {
		return token
	}

	return ""
}
