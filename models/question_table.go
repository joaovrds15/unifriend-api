package models

import "gorm.io/gorm"

type QuestionTable struct {
	gorm.Model
	Text    string `json:"text" gorm:"size:255;not null"`
	Quiz_id uint
	Quiz    QuizTable
	Options []OptionTable `gorm:"foreignKey:QuestionID"`
}
