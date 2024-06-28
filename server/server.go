package server

import (
	"log"
	"net/http"

	"github.com/google/uuid"
	"nhooyr.io/websocket"
)

type GameServer struct {

	// serveMux routes the various endpoints to the appropriate handler.
	serveMux http.ServeMux

	clients    map[uuid.UUID]*Client
	rooms      map[uuid.UUID]*Room
	register   chan *Client
	unregister chan *Client
	broadcast  chan []byte
}

const (
	roomCapacity = 2
)

func NewGameServer() *GameServer {
	gs := &GameServer{
		clients:    make(map[uuid.UUID]*Client),
		rooms:      make(map[uuid.UUID]*Room),
		register:   make(chan *Client),
		unregister: make(chan *Client),
		broadcast:  make(chan []byte),
	}

	go gs.RunServer()

	gs.serveMux.HandleFunc("/ws", gs.ServeGS)

	return gs
}

func (gs *GameServer) RunServer() {
	for {
		select {
		case client := <-gs.register:
			log.Printf("Client %s registered.", client.ClientId.String())
			gs.clients[client.ClientId] = client

		case client := <-gs.unregister:
			log.Printf("Client %s unregistered.", client.ClientId.String())
			delete(gs.clients, client.ClientId)

		case message := <-gs.broadcast:
			for _, client := range gs.clients {
				client.send <- message
			}
		}
	}
}

// ServeGS handles websocket requests from clients
func (gs *GameServer) ServeGS(w http.ResponseWriter, r *http.Request) {
	name := r.URL.Query().Get("name")
	if len(name) < 1 {
		log.Println("url param 'name' is missing.")
		return
	}

	conn, err := websocket.Accept(w, r, nil)
	if err != nil {
		log.Println("websocket upgrade error: ", err)
		return
	}

	client := NewClient(conn, name, gs)
	gs.register <- client

	readErr := client.ReadPump(r.Context())
	writeErr := client.WritePump(r.Context())

	select {
	case err := <-readErr:
		log.Println("read pump error: ", err)
	case err := <-writeErr:
		log.Println("write pump error: ", err)
	}
}

func (gs *GameServer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	gs.serveMux.ServeHTTP(w, r)
}

func (gs *GameServer) FindRoomById(id uuid.UUID) *Room {
	return gs.rooms[id]
}

func (gs *GameServer) FindClientById(id uuid.UUID) *Client {
	return gs.clients[id]
}

func (gs *GameServer) CreateRoom(client *Client) *Room {
	room := NewRoom(roomCapacity)
	go room.RunRoom()
	gs.rooms[room.RoomId] = room

	return room
}
