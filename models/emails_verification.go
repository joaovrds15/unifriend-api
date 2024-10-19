package models

import (
	"os"
	"strconv"
	"time"
)

type EmailsVerification struct {
	ID               uint      `json:"id" gorm:"primaryKey"`
	Email            string    `gorm:"size:100;not null"`
	VerificationCode int       `gorm:"not null"`
	CreatedAt        time.Time `gorm:"autoCreateTime"`
	Expiration       time.Time `gorm:"not null"`
}

var emailVerificationCodeLifespan int

func init() {
	verificationLifespan := os.Getenv("VERIFICATION_CODE_LIFESPAN_MINUTES")
	lifespan, err := strconv.Atoi(verificationLifespan)
	if err != nil {
		lifespan = 5
	}

	emailVerificationCodeLifespan = lifespan
}

func SaveVerificationCode(email string, code int) error {
	ev := EmailsVerification{
		Email:            email,
		VerificationCode: code,
		Expiration:       time.Now().Add(time.Duration(emailVerificationCodeLifespan) * time.Minute).UTC().Truncate(time.Second),
	}

	err := DB.Save(&ev).Error

	if err != nil {
		return err
	}

	return nil
}

func GetLastetVerificationCodeEmail(email string) (int, error) {
	var verificationCode EmailsVerification

	DB.Where("email = ?", email).Find(&verificationCode).Order("created_at")

	return verificationCode.VerificationCode, nil
}

func HasValidExpirationCode(email string) bool {
	var count int64

	fiveMinutesAgo := time.Now().Add(time.Duration(-emailVerificationCodeLifespan) * time.Minute)

	if err := DB.Where("email = ?", email).Where("expiration > ?", fiveMinutesAgo).Count(&count).Error; err != nil {
		return false
	}

	return count > 0
}

func CanSendVerificationCode(email string) bool {
	return !HasValidExpirationCode(email)
}
