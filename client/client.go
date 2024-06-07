package client

import "nhooyr.io/websocket"

// Client represents a single player.
type Client struct {

	// socket is the web socket for this client.
	socket *websocket.Conn

	// receive is a channel to receive messages from other clients
	receive chan []byte

	// room is the Room the client is in
	room *Room
}

func (c *Client) Read() {
	defer c.socket.Close()
	c.socket.Reader()
}
