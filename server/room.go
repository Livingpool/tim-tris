package server

import (
	"fmt"
)

type Room struct {
	Name     string
	capacity int

	//  holds all clients in this room.
	clients map[string]*Client

	// join is a channel for  wishing to join the room.
	join chan *Client

	// leave is a channel for  wishing to leave the room.
	leave chan *Client

	// broadcast holds incoming messages that should be forwarded to other
	broadcast chan *Message
}

func NewRoom(capacity int, name string) *Room {
	return &Room{
		Name:      name,
		capacity:  capacity,
		clients:   make(map[string]*Client),
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
			room.clients[client.Name] = client

		case client := <-room.leave:
			delete(room.clients, client.Name)

		case msg := <-room.broadcast:
			room.broadcastToClientsInRoom(msg.Encode())
		}
	}
}

func (room *Room) notifyClientJoined(client *Client) {
	message := Message{
		Action:  SendMessageAction,
		Target:  room.Name,
		Message: fmt.Sprintf("%s joined the room.", client.Name),
	}

	room.broadcastToClientsInRoom(message.Encode())
}

func (room *Room) broadcastToClientsInRoom(message []byte) {
	for _, client := range room.clients {
		client.send <- message
	}
}
