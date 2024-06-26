package models

type QuizTable struct {
	ID          uint            `json:"id" gorm:"primaryKey"`
	Title       string          `json:"title" gorm:"size:255;not null;unique"`
	Description string          `json:"description" gorm:"size:255;not null"`
	Questions   []QuestionTable `gorm:"foreignKey:Quiz_id"`
}

func GetQuizByID(id uint) (QuizTable, error) {

	var quiz QuizTable

	if err := DB.First(&quiz, id).Error; err != nil {
		return quiz, err
	}

	return quiz, nil

}
