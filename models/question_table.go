package models

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
