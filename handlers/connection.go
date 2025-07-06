package handlers

import (
	"net/http"
	"strconv"
	"time"
	"unifriend-api/models"

	"github.com/gin-gonic/gin"
)

type ConnectionRequestResponse struct {
    ID               uint      `json:"id"`
    RequestingUserID uint      `json:"requesting_user_id"`
    Status           int       `json:"status"`
    CreatedAt        string `json:"created_at"`
    RequestingUser   struct {
        UserId            uint   `json:"user_id"`
		ProfilePictureURL string `json:"profile_picture_url"`
		Name              string `json:"name"`
    } `json:"requesting_user"`
}

func CreateConnectionRequest (c *gin.Context) {
	requestedUserId := c.Param("user_id")
	requestedUserIdConverted, err := strconv.ParseUint(requestedUserId, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID format"})
		return
	}
	requestedUserIdUint32 := uint(requestedUserIdConverted)

	requestingUserId, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User ID not found in token"})
		return
	}

	requestingUserIdUint, ok := requestingUserId.(uint)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Invalid User ID format"})
		return
	}

	if ! models.ValidConnectionRequest(requestingUserIdUint, requestedUserIdUint32) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "you can not make this connection request"})
		return
	}
	
	connectionRequest := models.ConnectionRequest{
		RequestingUserID: requestingUserIdUint,
		RequestedUserID: requestedUserIdUint32,
	}

	if err := models.DB.Create(&connectionRequest).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "something went wrong"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"message": "Connection request created"})
}

func AcceptConnectionRequest(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User ID not found in token"})
		return
	}

	userIDUint, ok := userID.(uint)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Invalid User ID format"})
		return
	}

	connectionRequestId := c.Param("request_id")
	connectionRequestIdConverted, err := strconv.ParseUint(connectionRequestId, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID format"})
		return
	}

	connectionRequestUint32 := uint(connectionRequestIdConverted)

	connectionRequest, err := models.GetConnectionRequestById(connectionRequestUint32, userIDUint)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	connectionRequest.AnswerAt = time.Now()
	connectionRequest.Status = 1

	if err := models.DB.Save(&connectionRequest).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to accept connection request"})
		return
	}

	connection := models.Connection{
		UserAID: connectionRequest.RequestingUserID,
		UserBID: userIDUint,
		ConnectionRequestID: connectionRequest.ID,
	}

	if err := models.DB.Save(&connection).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to accept connection request"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":    "Connection request accepted",
	})
}

func GetConnectionRequests(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User ID not found in token"})
		return
	}

	userIDUint, ok := userID.(uint)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Invalid User ID format"})
		return
	}

	connectionRequests, err := models.GetConnectionRequests(userIDUint)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	var responses []ConnectionRequestResponse
	for _, req := range connectionRequests {
		responses = append(responses, ConnectionRequestResponse{
			ID:               req.ID,
			RequestingUserID: req.RequestingUserID,
			CreatedAt:        req.CreatedAt.Format("2006-01-02 15:04"),
			Status: 		  req.Status,
			RequestingUser: struct {
				UserId            uint   `json:"user_id"`
				ProfilePictureURL string `json:"profile_picture_url"`
				Name              string `json:"name"`
			}{
				UserId:            req.RequestingUser.ID,
				ProfilePictureURL: req.RequestingUser.ProfilePictureURL,
				Name:              req.RequestingUser.Name,
			},
		})
	}

	c.JSON(http.StatusOK, gin.H{"data": responses})
}

func RejectConnectionRequest(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User ID not found in token"})
		return
	}

	userIDUint, ok := userID.(uint)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Invalid User ID format"})
		return
	}

	connectionRequestId := c.Param("request_id")
	connectionRequestIdConverted, err := strconv.ParseUint(connectionRequestId, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request ID format"})
		return
	}

	connectionRequestUint32 := uint(connectionRequestIdConverted)

	connectionRequest, err := models.GetConnectionRequestById(connectionRequestUint32, userIDUint)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	connectionRequest.AnswerAt = time.Now()
	connectionRequest.Status = 0

	if err := models.DB.Save(&connectionRequest).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to accept connection request"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":    "Connection request rejected",
	})
}

func DeleteConnection(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User ID not found in token"})
		return
	}

	userIDUint, ok := userID.(uint)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Invalid User ID format"})
		return
	}

	requestId := c.Param("connection_id")
	connectionIdConverted, err := strconv.ParseUint(requestId, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID format"})
		return
	}

	connectionUint32 := uint(connectionIdConverted)

	err = models.DeleteConnection(connectionUint32, userIDUint)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":    "connection deleted",
	})
}