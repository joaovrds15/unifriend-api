package models

import "gorm.io/gorm"

type QuizTable struct {
	gorm.Model
	Title       string          `json:"title" gorm:"size:255;not null;unique"`
	Description string          `json:"description" gorm:"size:255;not null"`
	Questions   []QuestionTable `gorm:"foreignKey:Quiz_id"`
}
