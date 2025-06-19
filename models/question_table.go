package models

import (
	"strings"
)

type QuestionTable struct {
	ID      uint   `json:"id" gorm:"primaryKey"`
	Text    string `json:"text" gorm:"size:255;not null"`
	Quiz_id uint   `json:"quizId"`
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

func GetQuestions() ([]QuestionTable, error) {

	var questions []QuestionTable

	if err := DB.Model(&QuestionTable{}).Preload("Options").Find(&questions).Error; err != nil {
		return questions, err
	}

	return questions, nil

}

func GetQuestionsAndOptions(userAnswers []OptionTable) ([]OptionTable, error) {

	var optionsAndQuestions []OptionTable

	var queryConditions strings.Builder
	var queryArgs []interface{}

	for i, answer := range userAnswers  {
		if i > 0 {
			queryConditions.WriteString(" OR ")
		}
		queryConditions.WriteString("(question_id = ? AND id = ?)")
		queryArgs = append(queryArgs, answer.QuestionID, answer.ID)
	}

	if err := DB.Model(&OptionTable{}).Where(queryConditions.String(), queryArgs...).Find(&optionsAndQuestions).Error; err != nil {
		return optionsAndQuestions, err
	}

	return optionsAndQuestions, nil

}
