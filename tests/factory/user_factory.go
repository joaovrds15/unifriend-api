package factory

import (
	"fmt"
	"os"
	"unifriend-api/models"
	"unifriend-api/utils/token"

	"github.com/go-faker/faker/v4"
	"golang.org/x/exp/rand"
)

func UserFactory() models.User {
	user := models.User{
		Email:             faker.Email(),
		Password:          faker.Password(),
		Name:              faker.Name(),
		PhoneNumber:       generateBrazilianPhoneNumber(),
		ProfilePictureURL: faker.URL(),
		IsAdmin:           false,
		MajorID:           MajorFactory().ID,
		Images:            []models.UsersImages{UsersImagesFactory()},
	}

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

func generateBrazilianPhoneNumber() string {

	// Generate a random two-digit area code (10 to 99)
	areaCode := rand.Intn(90) + 10

	// Brazilian mobile numbers always start with 9
	firstDigit := 9

	// Generate the remaining 8 digits of the phone number
	remainingDigits := rand.Intn(90000000) + 10000000

	// Concatenate all digits into a single string
	phoneNumber := fmt.Sprintf("%02d%d%d", areaCode, firstDigit, remainingDigits)
	return phoneNumber
}
