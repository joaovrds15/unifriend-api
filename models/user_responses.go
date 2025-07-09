package models

import (
	"strings"
)

type UserResponse struct {
	ID         uint `json:"id" gorm:"primaryKey"`
	QuestionID uint
	Question   QuestionTable
	OptionID   uint
	Option     OptionTable
	UserID     uint
	User       User
	HasConnection               bool `gorm:"->;-:migration;column:has_connection"`
    HasPendingConnectionRequest bool `gorm:"->;-:migration;column:has_pending_connection_request"`
}

type MatchingUserResponse struct {
    UserResponse                `gorm:"embedded"`
    HasConnection               bool              `gorm:"column:has_connection"`
    HasPendingConnectionRequest bool              `gorm:"column:has_pending_connection_request"`
    ConnectionRequest           ConnectionRequest `gorm:"embedded;embeddedPrefix:cr_"`
}

func (u *UserResponse) SaveUserResponse() {
	DB.Save(&u)
}

func SaveUserResponses(userResponses []UserResponse) error {
	if len(userResponses) == 0 {
		return nil
	}

	if err := DB.Create(&userResponses).Error; err != nil {
		return err
	}

	return nil
}

func GetUserResponsesByUserID(userId uint) ([]UserResponse, error) {
	var userResponses []UserResponse

	if err := DB.Where("user_id = ?", userId).Find(&userResponses).Error; err != nil {
		return userResponses, err
	}

	return userResponses, nil
}

func HasUserAlreadyTakenQuiz(userID uint) (bool, error) {
	var count int64
	err := DB.Model(&UserResponse{}).
		Where("user_responses.user_id = ?", userID).
		Count(&count).Error

	if err != nil {
		return false, err
	}

	return count > 0, nil
}

func GetMatchingResponsesFromOtherUsers(currentUserID uint, currentUserAnswers []UserResponse) ([]MatchingUserResponse, error) {
    var matchingResponses []MatchingUserResponse

    if len(currentUserAnswers) == 0 {
        return matchingResponses, nil
    }

    var queryConditions strings.Builder
    var queryArgs []interface{}

    for i, answer := range currentUserAnswers {
        if i > 0 {
            queryConditions.WriteString(" OR ")
        }
        queryConditions.WriteString("(user_responses.question_id = ? AND user_responses.option_id = ?)")
        queryArgs = append(queryArgs, answer.QuestionID, answer.OptionID)
    }

    connectionSubquery := `
        LEFT JOIN connection_requests cr ON 
        (cr.requesting_user_id = ? AND cr.requested_user_id= user_responses.user_id) OR 
        (cr.requesting_user_id = user_responses.user_id AND cr.requested_user_id = ?)`

    selectClause := `
        user_responses.*,
        cr.id as cr_id,
        cr.requesting_user_id as cr_requesting_user_id,
        cr.requested_user_id as cr_requested_user_id,
        cr.status as cr_status,
        CASE WHEN cr.id IS NOT NULL AND cr.status = 1 THEN TRUE ELSE FALSE END as has_connection,
        CASE WHEN cr.id IS NOT NULL AND cr.status = 2 THEN TRUE ELSE FALSE END as has_pending_connection_request`

    err := DB.Table("user_responses").
        Select(selectClause).
        Joins("JOIN users ON users.id = user_responses.user_id").
        Joins(connectionSubquery, currentUserID, currentUserID).
        Preload("User").
        Where("users.deleted_at IS NULL AND users.status = 1").
        Where("user_responses.user_id != ?", currentUserID).
        Where(queryConditions.String(), queryArgs...).
        Find(&matchingResponses).Error

    if err != nil {
        return nil, err
    }

    return matchingResponses, nil
}