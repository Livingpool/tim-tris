package client

import (
	"context"
	"io"
	"net/http"
	"time"

	"github.com/google/uuid"
	"nhooyr.io/websocket"
)

// Client represents a single player.
type Client struct {
	clientId string

	// socket is the web socket for this client.
	socket *websocket.Conn

	// receive is a channel to receive messages from other clients
	receive chan []byte

	// room is the Room the client is in
	room *Room
}

// CreateRoom() creates a room with the player being the first player
func (c *Client) CreateRoom(w http.ResponseWriter, r *http.Request) *Room {
	newRoom := NewRoom()
	newClient := &Client{
		clientId: uuid.New().String(),
		room:     newRoom,
	}

	newRoom.clients[newClient] = true

	return newRoom
}

// Subscribe() tries to subscribe to a room with id

// Read() reads from the receive channel
func (c *Client) Read(ctx context.Context) error {
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	for {
		typ, r, err := c.socket.Reader(ctx)
		if err != nil {
			return err
		}

		msg, err := io.ReadAll(r)
		if err != nil {
			return err
		}

		c.room.forward <- msg
		c.socket.Write()
	}
}

func writeTimeout(ctx context.Context, timeout time.Duration, c *websocket.Conn, msg []byte) error {
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	return c.Write(ctx, websocket.MessageText, msg)
}
