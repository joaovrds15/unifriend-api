package models

import "gorm.io/gorm"

type QuestionTable struct {
	gorm.Model
	Text    string `json:"text" gorm:"size:255;not null"`
	Quiz_id uint
	Quiz    QuizTable
	Options []OptionTable `gorm:"foreignKey:QuestionID"`
}

func GetQuestionByID(id uint) (QuestionTable, error) {

	var question QuestionTable

	if err := DB.First(&question, id).Error; err != nil {
		return question, err
	}

	return question, nil

}
