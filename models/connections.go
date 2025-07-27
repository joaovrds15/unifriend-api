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
    ConnectionRequestID uint `gorm:"not null;index;column:connection_request_id" json:"connection_request_id"`
    ConnectionRequest ConnectionRequest `gorm:"foreignKey:ConnectionRequestID" json:"-"`
    CreatedAt time.Time `gorm:"precision:3;autoCreateTime" json:"created_at"`
}

type ConnectionWithUser struct {
    Connection          `gorm:"embedded" json:"connections"`
    UserID            uint   `json:"user_id"`
    Name              string `json:"name"`
    ProfilePictureURL string `json:"profile_picture_url"`
    MessageID        uint      `json:"message_id"`
    Content          string    `json:"content"`
    Created          time.Time `json:"created"`
    ReadAt          *time.Time `json:"read_at"`
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

func GetConnections(userId uint) ([]ConnectionWithUser, error) {
    var results []ConnectionWithUser

    err := DB.Table("connections as c").
        Select(`c.id, c.user_a, c.user_b, c.created_at, c.connection_request_id, 
                u.id as user_id, u.name, u.profile_picture_url, 
                m.id as message_id, m.content, m.created_at as created, m.read_at`).
        Joins(`JOIN (
                SELECT connection_id, MAX(created_at) AS latest_message_time 
                FROM messages 
                GROUP BY connection_id
              ) latest ON latest.connection_id = c.id`).
        Joins(`JOIN messages m ON m.connection_id = latest.connection_id 
               AND m.created_at = latest.latest_message_time`).
        Joins(`JOIN users u ON u.id = CASE WHEN c.user_a = ? THEN c.user_b ELSE c.user_a END`, userId).
        Where("c.user_a = ? OR c.user_b = ?", userId, userId).
        Where("u.status = 1 AND u.deleted_at IS NULL").
        Find(&results).Error

    return results, err
}