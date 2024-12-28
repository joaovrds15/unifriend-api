package token

import (
	"os"
	"strconv"
	"time"

	jwt "github.com/dgrijalva/jwt-go"
)

type RegistrationToken struct {
	Email string `json:"email"`
}

func (r RegistrationToken) GenerateToken() (string, error) {

	token_lifespan, err := strconv.Atoi(os.Getenv("TOKEN_REGISTRATION_HOUR_LIFESPAN"))

	if err != nil {
		return "", err
	}

	claims := jwt.MapClaims{}
	claims["authorized"] = true
	claims["email"] = r.Email
	claims["exp"] = time.Now().Add(time.Hour * time.Duration(token_lifespan)).Unix()
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	return token.SignedString([]byte(os.Getenv("API_SECRET")))

}
