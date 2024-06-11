package models

import (
	"errors"
	"strings"
	"unifriend-api/utils/token"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type User struct {
	ID                uint   `json:"id" gorm:"primaryKey"`
	Email             string `gorm:"size:100;unique;not null"`
	Password          string `json:"password" gorm:"size:100;not null"`
	Name              string `gorm:"size:100;not null"`
	ProfilePictureURL string `gorm:"size:255"`
	IsAdmin           bool   `gorm:"default:false"`
	MajorID           uint
	Major             Major
}

func GetUserByID(uid uint) (User, error) {

	var u User

	if err := DB.First(&u, uid).Error; err != nil {
		return u, errors.New("User not found")
	}

	u.PrepareGive()

	return u, nil

}

func UsernameAlreadyUsed(email string) bool {
	var count int64
	DB.Model(&User{}).Where("email = ?", email).Count(&count)
	return count > 0
}

func (u *User) PrepareGive() {
	u.Password = ""
}

func (u *User) SaveUser() (*User, error) {
	err := DB.Create(&u).Error
	if err != nil {
		return &User{}, err
	}
	return u, nil
}

func (u *User) BeforeSave(DB *gorm.DB) error {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(u.Password), bcrypt.DefaultCost)

	if err != nil {
		return err
	}
	u.Password = string(hashedPassword)
	u.Email = strings.TrimSpace(u.Email)

	return nil
}

func VerifyPassword(password, hashedPassword string) error {
	return bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))
}

func LoginCheck(email string, password string) (string, error) {
	var err error

	u := User{}

	err = DB.Model(User{}).Where("email = ?", email).Take(&u).Error

	if err != nil {
		return "", err
	}

	err = VerifyPassword(password, u.Password)

	if err != nil && err == bcrypt.ErrMismatchedHashAndPassword {
		return "", err
	}

	token, err := token.GenerateToken(u.ID)

	if err != nil {
		return "", err
	}

	return token, nil

}
