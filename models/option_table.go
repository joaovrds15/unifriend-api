package models

type OptionTable struct {
	ID         uint   `json:"id" gorm:"primaryKey"`
	Text       string `json:"text" gorm:"size:255;not null"`
	QuestionID uint
	QuestionTable
	UserResponses []UserResponse `gorm:"foreignKey:OptionID"`
}
