package models

import "strings"

type UserResponse struct {
	ID         uint `json:"id" gorm:"primaryKey"`
	QuestionID uint
	Question   QuestionTable
	OptionID   uint
	Option     OptionTable
	UserID     uint
	User       User
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

func GetMatchingResponsesFromOtherUsers(currentUserID uint, currentUserAnswers []UserResponse) ([]UserResponse, error) {
	var matchingResponses []UserResponse

	if len(currentUserAnswers) == 0 {
		return matchingResponses, nil
	}

	var queryConditions strings.Builder
	var queryArgs []interface{}

	for i, answer := range currentUserAnswers {
		if i > 0 {
			queryConditions.WriteString(" OR ")
		}
		queryConditions.WriteString("(question_id = ? AND option_id = ?)")
		queryArgs = append(queryArgs, answer.QuestionID, answer.OptionID)
	}

	err := DB.Preload("User").
		Joins("JOIN users ON users.id = user_responses.user_id").
		Where("users.deleted_at IS NULL AND users.status = 1").
		Where("user_responses.user_id != ?", currentUserID).
		Where(queryConditions.String(), queryArgs...).
		Find(&matchingResponses).Error

	if err != nil {
		return nil, err
	}

	return matchingResponses, nil
}
