package models

import (
	"errors"
	"strings"
	"time"
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
	PhoneNumber       string `gorm:"size:20;not null"`
	MajorID           uint
	Major             Major	`gorm:"foreignKey:MajorID"`
	Status			  int  `gorm:"default:1"`
	Images            []UsersImages `gorm:"foreignKey:UserID"`
	UserResponses     []UserResponse `gorm:"foreignKey:UserID"`
	DeletedAt        time.Time `gorm:"default:NULL"`
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

func PhoneNumberAlreadyUsed(phoneNumber string) bool {
	var count int64
	DB.Model(&User{}).Where("phone_number = ?", phoneNumber).Count(&count)
	return count > 0
}

func (u *User) PrepareGive() {
	u.Password = ""
}

func (u *User) SaveUser() (*User, error) {
	err := DB.Save(&u).Error
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

func LoginCheck(email string, password string) (string, User) {
	var err error

	u := User{}

	err = DB.Model(User{}).Where("email = ?", email).Where("status = 1 AND deleted_at IS NULL").Take(&u).Error

	if err != nil {
		return "", User{}
	}

	err = VerifyPassword(password, u.Password)

	if err != nil && err == bcrypt.ErrMismatchedHashAndPassword {
		return "", User{}
	}

	token, err := token.GenerateToken(u.ID)

	if err != nil {
		return "", User{}
	}

	return token, u

}

func (u *User) DeleteUser() error {
    return DB.Transaction(func(tx *gorm.DB) error {
        updates := map[string]interface{}{
            "Status":    0,
            "DeletedAt": time.Now().Format("2006-01-02 15:04:05"),
		}

        if err := tx.Model(&u).Updates(updates).Error; err != nil {
            return err
        }

        return nil
    })
}