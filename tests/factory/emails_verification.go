package factory

import (
	"os"
	"strconv"
	"time"
	"unifriend-api/models"

	"github.com/go-faker/faker/v4"
	"golang.org/x/exp/rand"
)

var emailVerificationCodeLifespan int

func init() {
	verificationLifespan := os.Getenv("VERIFICATION_CODE_LIFESPAN_MINUTES")
	lifespan, err := strconv.Atoi(verificationLifespan)
	if err != nil {
		lifespan = 5
	}

	emailVerificationCodeLifespan = lifespan
}

func EmailsVerificationFactory() models.EmailsVerification {
	rand.Seed(uint64(time.Now().UnixNano()))
	verificationCode := rand.Intn(900000) + 100000

	emailVerification := models.EmailsVerification{
		Email:            faker.Email(),
		VerificationCode: verificationCode,
		CreatedAt:        time.Now().UTC(),
		Expiration:       time.Now().Add(time.Duration(emailVerificationCodeLifespan) * time.Minute).UTC().Truncate(time.Second),
	}

	return emailVerification
}
