package server

import (
	"fmt"

	"github.com/google/uuid"
)

type Room struct {
	RoomId uuid.UUID

	capacity int

	//  holds all clients in this room.
	clients map[*Client]bool

	// join is a channel for  wishing to join the room.
	join chan *Client

	// leave is a channel for  wishing to leave the room.
	leave chan *Client

	// broadcast holds incoming messages that should be forwarded to other
	broadcast chan *Message
}

func NewRoom(capacity int) *Room {
	return &Room{
		RoomId:    uuid.New(),
		capacity:  capacity,
		clients:   make(map[*Client]bool),
		join:      make(chan *Client),
		leave:     make(chan *Client),
		broadcast: make(chan *Message),
	}
}

func (room *Room) RunRoom() {
	for {
		select {
		case client := <-room.join:
			room.notifyClientJoined(client)
			room.clients[client] = true

		case client := <-room.leave:
			delete(room.clients, client)

		case msg := <-room.broadcast:
			room.broadcastToClientsInRoom(msg.Encode())
		}
	}
}

func (room *Room) notifyClientJoined(client *Client) {
	message := Message{
		Action:  SendMessageAction,
		Target:  room,
		Message: fmt.Sprintf("%s joined the room.", client.Name),
	}

	room.broadcastToClientsInRoom(message.Encode())
}

func (room *Room) broadcastToClientsInRoom(message []byte) {
	for client := range room.clients {
		client.send <- message
	}
}
