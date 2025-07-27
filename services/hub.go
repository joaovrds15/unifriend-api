package services

import (
	"log"
	"unifriend-api/models"
)

type Hub struct {
    clients map[uint]*Client
    broadcast chan *models.Message
    register chan *Client
    unregister chan *Client
}

func NewHub() *Hub {
    return &Hub{
        broadcast:  make(chan *models.Message),
        register:   make(chan *Client),
        unregister: make(chan *Client),
        clients:    make(map[uint]*Client),
    }
}

func (h *Hub) Broadcast(message *models.Message) {
    h.broadcast <- message
}


func (h *Hub) Run() {
    for {
        select {
        case client := <-h.register:
            h.clients[client.userID] = client
        case client := <-h.unregister:
            if _, ok := h.clients[client.userID]; ok {
                delete(h.clients, client.userID)
                close(client.send)
            }
        case message := <-h.broadcast:
            var connection models.Connection
            if err := models.DB.First(&connection, message.ConnectionID).Error; err != nil {
                log.Printf("error finding connection: %v", err)
                continue
            }

            var recipientID uint
            if message.SenderID == connection.UserAID {
                recipientID = connection.UserBID
            } else {
                recipientID = connection.UserAID
            }

            if client, ok := h.clients[recipientID]; ok {
                select {
                case client.send <- message:
                default:
                    close(client.send)
                    delete(h.clients, client.userID)
                }
            }
        }
    }
}