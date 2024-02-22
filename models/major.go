package models

import "gorm.io/gorm"

type Major struct {
	gorm.Model
	Name string `json:"name" gorm:"size:255;not null;unique"`
}
