package models

type OptionTable struct {
	ID         uint   `json:"id" gorm:"primaryKey"`
	Text       string `json:"text" gorm:"size:255;not null"`
	QuestionID uint
	QuestionTable
	UserResponses []UserResponse `gorm:"foreignKey:OptionID"`
}

func GetOptionByID(id uint) (OptionTable, error) {

	var option OptionTable

	if err := DB.First(&option, id).Error; err != nil {
		return option, err
	}

	return option, nil

}
