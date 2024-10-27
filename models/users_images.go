package models

import "time"

type UsersImages struct {
	ID        uint   `json:"id" gorm:"primaryKey"`
	ImageUrl  string `json:"image_url" gorm:"size:255;not null"`
	UserID    uint   `json:"user_id" gorm:"not null"`
	User      User
	CreatedAt time.Time `gorm:"autoCreateTime"`
	UpdatedAt time.Time `gorm:"autoUpdateTime"`
}
