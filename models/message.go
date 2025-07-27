package models

import "time"

type Message struct {
    ID           uint      `gorm:"primaryKey;autoIncrement" json:"id"`
    ConnectionID uint      `gorm:"not null;index" json:"connection_id"`
    Connection   Connection `gorm:"foreignKey:ConnectionID;constraint:OnDelete:CASCADE" json:"-"`
    SenderID     uint      `gorm:"not null;index" json:"sender_id"`
    Sender       User      `gorm:"foreignKey:SenderID;constraint:OnDelete:CASCADE" json:"-"`
	ReceiverID   uint      `gorm:"not null;index" json:"receiver_id"`
    Receiver     User      `gorm:"foreignKey:ReceiverID;constraint:OnDelete:CASCADE" json:"-"`
    Content      string    `gorm:"type:text;not null" json:"content"`
    CreatedAt    time.Time `gorm:"autoCreateTime" json:"created_at"`
    ReadAt       *time.Time `json:"read_at"`
}

func GetMessages(connectionId uint) ([]Message, error) {
    var results []Message

    err := DB.Where("connection_id = ?", connectionId).Order("created_at asc").Find(&results).Error;
    return results, err
}