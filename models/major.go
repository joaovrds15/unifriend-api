package models

type Major struct {
	ID   uint   `json:"id" gorm:"primaryKey"`
	Name string `json:"name" gorm:"size:255;not null;unique"`
}

func GetMajors() ([]Major, error) {

	var majors []Major

	if err := DB.Find(&majors).Error; err != nil {
		return majors, err
	}

	return majors, nil

}
