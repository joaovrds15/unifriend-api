package models

type OptionTable struct {
	ID            uint          `json:"option_id" gorm:"primaryKey"`
	Text          string        `json:"text" gorm:"size:255;not null"`
	QuestionID    uint          `json:"question_id"`
	QuestionTable QuestionTable `gorm:"foreignKey:QuestionID"`
}

func GetOptionByID(id uint) (OptionTable, error) {

	var option OptionTable

	if err := DB.First(&option, id).Error; err != nil {
		return option, err
	}

	return option, nil

}

func GetOptions() ([]OptionTable, error) {

	var options []OptionTable

	if err := DB.Find(&options).Error; err != nil {
		return options, err
	}

	return options, nil

}
