package models

import (
	"errors"
	"time"

	"gorm.io/gorm"
)

type Connection struct {
    ID        uint      `gorm:"primaryKey;autoIncrement" json:"id"`
    UserAID   uint      `gorm:"not null;index;column:user_a" json:"user_a"`
    UserBID   uint      `gorm:"not null;index;column:user_b" json:"user_b"`
    UserA     User      `gorm:"foreignKey:UserAID;constraint:OnDelete:CASCADE" json:"-"`
    UserB     User      `gorm:"foreignKey:UserBID;constraint:OnDelete:CASCADE" json:"-"`
    ConnectionRequestID uint `gorm:"not null;index;column:connection_request_id" json:"-"`
    ConnectionRequest ConnectionRequest `gorm:"foreignKey:ConnectionRequestID" json:"-"`
    CreatedAt time.Time `gorm:"precision:3;autoCreateTime" json:"created_at"`
}


func DeleteConnection(connectionId uint, userId uint) error {
    return DB.Transaction(func(tx *gorm.DB) error {
        var connection Connection

        if err := tx.Where("id = ? AND (user_a = ? OR user_b = ?)", connectionId, userId, userId).First(&connection).Error; err != nil {
            if err == gorm.ErrRecordNotFound {
                return errors.New("connection not found or user does not have permission to delete it")
            }
            return err
        }

        if err := tx.Delete(&ConnectionRequest{}, connection.ConnectionRequestID).Error; err != nil {
            return err
        }

        if err := tx.Delete(&connection).Error; err != nil {
            return err
        }

        return nil
    })
}