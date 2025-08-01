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
	Verified		 bool 		`gorm:"default:false"`
	CreatedAt        time.Time `gorm:"autoCreateTime"`
	Expiration       time.Time `gorm:"not null"`
}

func SaveVerificationCode(email string, code int) (*EmailsVerification, error) {
	verificationLifespan := os.Getenv("VERIFICATION_CODE_LIFESPAN_MINUTES")
	lifespan, err := strconv.Atoi(verificationLifespan)
	if err != nil {
		lifespan = 5
	}

	ev := &EmailsVerification{
		Email:            email,
		VerificationCode: code,
		Expiration:       time.Now().Add(time.Duration(lifespan) * time.Minute).UTC().Truncate(time.Second),
	}

	err = DB.Save(&ev).Error

	if err != nil {
		return &EmailsVerification{}, err
	}

	return ev, nil
}

func GetLastetVerificationCodeEmail(email string) (*EmailsVerification, error) {
	var verificationCode EmailsVerification

	DB.Where("email = ?", email).Order("created_at DESC").Find(&verificationCode)

	return &verificationCode, nil
}

func HasValidVerificationCode(email string) bool {
	var count int64

	if err := DB.Where("email = ?", email).Where("expiration >= ?", time.Now()).Find(&EmailsVerification{}).Count(&count).Error; err != nil {
		return false
	}

	return count > 0
}

func CanSendVerificationCode(email string) bool {
	return !HasValidVerificationCode(email)
}
