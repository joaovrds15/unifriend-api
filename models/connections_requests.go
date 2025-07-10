package models

import (
	"time"
)

type ConnectionRequest struct {
    ID                uint       `gorm:"primaryKey;autoIncrement" json:"id"`
    RequestingUserID  uint       `gorm:"not null;index;column:requesting_user_id" json:"requesting_user_id"`
    RequestedUserID   uint       `gorm:"not null;index;column:requested_user_id" json:"requested_user_id"`
    RequestingUser    User       `gorm:"foreignKey:RequestingUserID;constraint:OnDelete:CASCADE" json:"-"`
    RequestedUser     User       `gorm:"foreignKey:RequestedUserID;constraint:OnDelete:CASCADE" json:"-"`
    Status            int        `gorm:"type:int;default:2" json:"status"`
    CreatedAt         time.Time  `gorm:"precision:3;autoCreateTime" json:"created_at"`
    AnswerAt        time.Time `gorm:"precision:3; default:NULL" json:"accepted_at,omitempty"`
}

const (
    StatusDenied   int = 0
    StatusAccepted int = 1
    StatusPending  int = 2
)

func ValidConnectionRequest(requestingUserId uint, requestedUserId uint) (bool) {
    var count int64
    err := DB.
    Table("users AS u2").
    Where("u2.id = ? AND u2.status = 1 AND u2.id <> ?", requestedUserId, requestingUserId).
    Where(`
        u2.id NOT IN (
            SELECT c.user_a FROM connections c WHERE c.user_b = ?
            UNION
            SELECT c.user_b FROM connections c WHERE c.user_a = ?
        )
    `, requestingUserId, requestingUserId).
    Where(`
        u2.id NOT IN (
            SELECT cr.requesting_user_id FROM connection_requests cr WHERE cr.requested_user_id = ? AND cr.status = 2
            UNION
            SELECT cr.requested_user_id FROM connection_requests cr WHERE cr.requesting_user_id = ? AND cr.status = 2
        )
    `, requestingUserId, requestingUserId).
    Count(&count).Error

    if err != nil {
        return false
    }

    return count == 1
}

func GetConnectionRequests(userId uint) ([]ConnectionRequest, error) {
    var connectionRequests []ConnectionRequest
    if err := DB.Preload("RequestingUser.UserResponses").Where("requested_user_id = ? AND status = ?", userId, StatusPending).Find(&connectionRequests).Error; err != nil {
        return connectionRequests, err
    }

    return connectionRequests, nil
}

func GetConnectionRequestById(requestId uint, userId uint) (ConnectionRequest, error) {
    var connectionRequest ConnectionRequest
    err := DB.Where(
        "id = ? AND status = ? AND requested_user_id = ?",
        requestId, StatusPending, userId,
    ).First(&connectionRequest).Error
    return connectionRequest, err
}