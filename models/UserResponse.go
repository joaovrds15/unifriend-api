package models

type UserResponse struct {
	ID         uint `json:"id" gorm:"primaryKey"`
	QuestionID uint
	Question   QuestionTable
	OptionID   uint
	Option     OptionTable
	UserID     uint
	User       User
}
