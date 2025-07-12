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
    AnswerAt        *time.Time `gorm:"precision:3; default:NULL" json:"accepted_at"`
}

type ConnectionRequestWithScoreResult struct {
    ConnectionRequest `gorm:"embedded"`
    Name string
    ProfilePictureUrl string
    Score int
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

func GetConnectionRequests(userId uint) ([]ConnectionRequestWithScoreResult, error) {
    var results []ConnectionRequestWithScoreResult
    err := DB.Table("connection_requests as cr").
        Select(`
            cr.id, cr.requesting_user_id, cr.created_at, cr.status,
            u.profile_picture_url, u.name,
            COUNT(ur2.option_id) as score
        `).
        Joins("JOIN users u ON u.id = cr.requesting_user_id").
        Joins(`
            LEFT JOIN (
                SELECT user_id, option_id FROM user_responses WHERE user_id = ?
            ) as ur ON 1=1
        `, userId).
        Joins(`
            LEFT JOIN user_responses as ur2 ON ur2.user_id = cr.requesting_user_id AND ur2.option_id = ur.option_id
        `).
        Where("cr.requested_user_id = ? AND cr.status = ?", userId, StatusPending).
        Group("cr.id, u.id").
        Order("cr.created_at desc").
        Scan(&results).Error

    return results, err
}

func GetConnectionRequestById(requestId uint, userId uint) (ConnectionRequest, error) {
    var connectionRequest ConnectionRequest
    err := DB.Where(
        "id = ? AND status = ? AND requested_user_id = ?",
        requestId, StatusPending, userId,
    ).First(&connectionRequest).Error
    return connectionRequest, err
}