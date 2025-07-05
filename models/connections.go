package models

import (
	"time"
)

type Connection struct {
    ID        uint      `gorm:"primaryKey;autoIncrement" json:"id"`
    UserAID   uint      `gorm:"not null;index;column:user_a" json:"user_a"`
    UserBID   uint      `gorm:"not null;index;column:user_b" json:"user_b"`
    UserA     User      `gorm:"foreignKey:UserAID;constraint:OnDelete:CASCADE" json:"-"`
    UserB     User      `gorm:"foreignKey:UserBID;constraint:OnDelete:CASCADE" json:"-"`
    CreatedAt time.Time `gorm:"precision:3;autoCreateTime" json:"created_at"`
}
