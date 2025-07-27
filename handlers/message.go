package handlers

import (
	"net/http"
	"strconv"
	"unifriend-api/models"
	"unifriend-api/services"

	"github.com/gin-gonic/gin"
)

type SendMessageInput struct {
    Content string `json:"content" binding:"required"`
}

type UserDTO struct {
    UserID            uint   `json:"user_id"`
    Name              string `json:"name"`
    ProfilePictureURL string `json:"profile_picture_url"`
}

type MessageWithUser struct {
    Messages []models.Message `json:"messages"`
    User UserDTO `json:"user"`
}

var hub *services.Hub

func SetHub(h *services.Hub) {
    hub = h
}

func HandleWebSocket(c *gin.Context, hub *services.Hub) {
    services.ServeWs(hub, c)
}

func SendMessage(c *gin.Context) {
    userID, exists := c.Get("user_id")
    if !exists {
        c.JSON(http.StatusUnauthorized, gin.H{"error": "User ID not found in token"})
        return
    }

    connectionID, err := strconv.ParseUint(c.Param("connection_id"), 10, 32)
    if err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid connection ID"})
        return
    }

    var input SendMessageInput
    if err := c.ShouldBindJSON(&input); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }

    var connection models.Connection
    if err := models.DB.First(&connection, uint(connectionID)).Error; err != nil {
        c.JSON(http.StatusNotFound, gin.H{"error": "Connection not found"})
        return
    }

    if connection.UserAID != userID.(uint) && connection.UserBID != userID.(uint) {
        c.JSON(http.StatusForbidden, gin.H{"error": "You are not part of this connection"})
        return
    }

    message := models.Message{
        ConnectionID: uint(connectionID),
        SenderID:     userID.(uint),
        Content:      input.Content,
    }

    if err := models.DB.Create(&message).Error; err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to send message"})
        return
    }

    if hub != nil {
        hub.Broadcast(&message)
    }

    c.JSON(http.StatusCreated, message)
}

func GetMessages(c *gin.Context) {
    userID, exists := c.Get("user_id")
    if !exists {
        c.JSON(http.StatusUnauthorized, gin.H{"error": "User ID not found in token"})
        return
    }

    connectionID, err := strconv.ParseUint(c.Param("connection_id"), 10, 32)
    if err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid connection ID"})
        return
    }

    var connection models.Connection
    if err := models.DB.First(&connection, uint(connectionID)).Error; err != nil {
        c.JSON(http.StatusNotFound, gin.H{"error": "Connection not found"})
        return
    }

    if connection.UserAID != userID.(uint) && connection.UserBID != userID.(uint) {
        c.JSON(http.StatusForbidden, gin.H{"error": "You are not part of this connection"})
        return
    }

    var messages []models.Message
    if messages, err = models.GetMessages(uint(connectionID)); err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve messages"})
        return
    }

    var otherUserID uint
    if connection.UserAID == userID.(uint) {
        otherUserID = connection.UserBID
    } else {
        otherUserID = connection.UserAID
    }
    
    var otherUser models.User
    if err := models.DB.First(&otherUser, otherUserID).Error; err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve user information"})
        return
    }

    response := MessageWithUser{
        Messages: messages,
        User: UserDTO{
            UserID: otherUser.ID,
            Name: otherUser.Name,
            ProfilePictureURL: otherUser.ProfilePictureURL,
        },
    }

    c.JSON(http.StatusOK, gin.H{"data": response})
}