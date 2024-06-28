package server

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"log"
	"time"

	"github.com/google/uuid"
	"nhooyr.io/websocket"
)

// Client represents a single player.
type Client struct {
	ClientId   uuid.UUID `json:"id"`
	Name       string    `json:"name"`
	conn       *websocket.Conn
	send       chan []byte
	gameServer *GameServer
	rooms      map[*Room]bool
}

func NewClient(conn *websocket.Conn, name string, gameServer *GameServer) *Client {
	return &Client{
		ClientId:   uuid.New(),
		Name:       name,
		conn:       conn,
		send:       make(chan []byte, 256),
		gameServer: gameServer,
		rooms:      make(map[*Room]bool),
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
		ticker := time.NewTicker(30 * time.Second)
		defer func() {
			ticker.Stop()
			client.conn.Close(websocket.StatusNormalClosure, "websocket closed")
		}()

		for {
			select {
			case message, ok := <-client.send:
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

			case <-ticker.C:
				if err := client.conn.Ping(ctx); err != nil {
					errChan <- err
					return
				}
			}
		}
	}()

	return errChan
}

func (client *Client) disconnect() {
	client.gameServer.unregister <- client
	for room := range client.rooms {
		room.leave <- client
	}
	close(client.send)
	client.conn.Close(websocket.StatusNormalClosure, "websocket closed")
}

// handleNewMessage handles the message based on those defined in message.go
func (client *Client) handleNewMessage(jsonMessage []byte) {

	var msg Message
	if err := json.Unmarshal(jsonMessage, &msg); err != nil {
		log.Printf("error on unmarshal JSON message %s\n", err)
		return
	}

	msg.Sender = client

	switch msg.Action {
	case SendMessageAction:
		roomId := msg.Target.RoomId
		if room := client.gameServer.FindRoomById(roomId); room != nil {
			room.broadcast <- &msg
		} else {
			log.Println("room Id not found on SendMessageAction")
		}

	case JoinRoomAction:
		roomId := msg.Target.RoomId
		room := client.gameServer.FindRoomById(roomId)
		client.joinRoom(room, client)

	case LeaveRoomAction:
		roomId := msg.Target.RoomId
		room := client.gameServer.FindRoomById(roomId)
		if room == nil {
			log.Println("room Id not found on LeaveRoomAction")
			return
		}

		delete(client.rooms, room)

		room.leave <- client
	}
}

func (client *Client) joinRoom(room *Room, sender *Client) {
	if room == nil {
		log.Println("Creating new room...")
		room = client.gameServer.CreateRoom(sender)
	}

	if !client.isInRoom(room) {
		client.rooms[room] = true
		room.join <- client
		log.Printf("Client %s joined room %s", client.ClientId, room.RoomId)
		client.notifyRoomJoined(room, sender)
	}
}

func (client *Client) isInRoom(room *Room) bool {
	if _, ok := client.rooms[room]; ok {
		return true
	}

	return false
}

func (client *Client) notifyRoomJoined(room *Room, sender *Client) {
	message := Message{
		Action: RoomJoinedAction,
		Target: room,
		Sender: sender,
	}

	client.send <- message.Encode()
}
