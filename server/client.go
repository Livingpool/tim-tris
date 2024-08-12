package server

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"log"
	"time"

	"github.com/coder/websocket"
)

// Client represents a single player.
type Client struct {
	Name       string `json:"name"`
	conn       *websocket.Conn
	send       chan []byte
	gameServer *GameServer
	rooms      map[string]*Room
}

func NewClient(conn *websocket.Conn, name string, gameServer *GameServer) *Client {
	return &Client{
		Name:       name,
		conn:       conn,
		send:       make(chan []byte, 256),
		gameServer: gameServer,
		rooms:      make(map[string]*Room),
	}
}

// ReadPump() reads messages from the websocket till the client disconnects.
func (client *Client) ReadPump(ctx context.Context) <-chan error {
	errChan := make(chan error)
	go func() {
		defer func() {
			client.disconnect()
		}()

		for {
			ctx, cancel := context.WithTimeout(ctx, 180*time.Second)
			defer cancel()

			_, r, err := client.conn.Reader(ctx)
			if err != nil {
				errChan <- err
				return
			}

			msg, err := io.ReadAll(r)
			if err != nil {
				errChan <- err
				return
			}
			client.handleNewMessage(msg)
		}

	}()
	return errChan
}

// WritePump() waits for messages from the send channel and forwards them to the client.
func (client *Client) WritePump(ctx context.Context) <-chan error {
	errChan := make(chan error)
	go func() {
		defer func() {
			client.conn.Close(websocket.StatusNormalClosure, "websocket closed")
		}()

		for {
			ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
			defer cancel()

			message, ok := <-client.send
			if !ok {
				// GameServer closed the channel.
				client.conn.Write(ctx, websocket.MessageText, []byte("websocket closed"))
				errChan <- errors.New("game server closed the websocket")
				return
			}

			w, err := client.conn.Writer(ctx, websocket.MessageText)
			if err != nil {
				errChan <- err
				return
			}
			w.Write(message)

			// Attach queued chat messages to the current websocket message.
			n := len(client.send)
			for range n {
				w.Write([]byte{'\n'})
				w.Write(<-client.send)
			}

			if err = w.Close(); err != nil {
				errChan <- err
				return
			}
		}
	}()

	return errChan
}

func (client *Client) disconnect() {
	client.gameServer.unregister <- client
	for _, room := range client.rooms {
		room.leave <- client
	}
	close(client.send)
	client.conn.Close(websocket.StatusNormalClosure, "websocket closed")
}

// handleNewMessage handles the message based on those defined in message.go
func (client *Client) handleNewMessage(jsonMessage []byte) {
	var msg Message
	if err := json.Unmarshal(jsonMessage, &msg); err != nil {
		client.gameServer.Logf("%v\n", err)
		return
	}

	msg.Sender = client.Name

	room, exists := client.gameServer.rooms[msg.Target]
	if !exists {
		client.gameServer.Logf("roomId not found")
		return
	}

	switch msg.Action {
	case SendMessageAction:
		room.broadcast <- &msg
		client.gameServer.Logf(msg.Message)

	case JoinRoomAction:
		client.joinRoom(room, client)

	case LeaveRoomAction:
		delete(client.rooms, room.Name)
		room.leave <- client
	}
}

func (client *Client) joinRoom(room *Room, sender *Client) {
	if !client.isInRoom(room) {
		client.rooms[room.Name] = room
		room.join <- client
		log.Printf("Client %s joined room %s\n", client.Name, room.Name)
		client.notifyRoomJoined(room, sender)
	}
}

func (client *Client) isInRoom(room *Room) bool {
	if _, ok := client.rooms[room.Name]; ok {
		return true
	}

	return false
}

func (client *Client) notifyRoomJoined(room *Room, sender *Client) {
	message := Message{
		Action: RoomJoinedAction,
		Target: room.Name,
		Sender: sender.Name,
	}

	client.send <- message.Encode()
}
