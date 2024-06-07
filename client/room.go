package client

import (
	"log"
	"net/http"

	"nhooyr.io/websocket"
)

type Room struct {
	roomId int

	// clients holds all clients in this room.
	clients map[*Client]bool

	// join is a channel for clients wishing to join the room.
	join chan *Client

	// leave is a channel for clients wishing to leave the room.
	leave chan *Client

	// forward holds incoming messages that should be forwarded to other clients
	forward chan []byte

	// logf controls where logs are sent.
	// Defaults to log.Printf.
	logf func(f string, v ...interface{})
}

const (
	socketBufferSize  = 1024
	messageBufferSize = 256
)

func NewRoom() *Room {
	return &Room{
		clients: make(map[*Client]bool),
		join:    make(chan *Client),
		leave:   make(chan *Client),
		forward: make(chan []byte),
		logf:    log.Printf,
	}
}

func (r *Room) Run() {
	for {
		select {
		case client := <-r.join:
			r.clients[client] = true
		case client := <-r.leave:
			delete(r.clients, client)
			close(r.leave)
		case msg := <-r.forward:
			for client := range r.clients {
				client.receive <- msg
			}
		}
	}
}

// Make
func (r *Room) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	c, err := websocket.Accept(w, req, &websocket.AcceptOptions{})
	if err != nil {
		r.logf("%v", err)
	}

	client := &Client{
		socket:  c,
		receive: make(chan []byte, messageBufferSize),
		room:    r,
	}

	r.join <- client
	defer func() { r.leave <- client }()

	go client.Write()
	client.Read()
}
