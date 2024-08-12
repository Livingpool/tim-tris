package server

import (
	"log"
	"net/http"

	"github.com/Livingpool/tim-tris/web"
	"github.com/coder/websocket"
)

type GameServer struct {

	// serveMux routes the various endpoints to the appropriate handler.
	serveMux http.ServeMux

	// Logf controls where logs are sent.
	// Defaults to log.Printf.
	Logf func(f string, v ...interface{})

	clients    map[string]*Client
	rooms      map[string]*Room
	register   chan *Client
	unregister chan *Client
	broadcast  chan []byte

	renderer web.TemplatesInterface
}

const (
	roomCapacity = 2
)

func NewGameServer() *GameServer {
	gs := &GameServer{
		Logf:       log.Printf,
		clients:    make(map[string]*Client),
		rooms:      make(map[string]*Room),
		register:   make(chan *Client),
		unregister: make(chan *Client),
		broadcast:  make(chan []byte),
		renderer:   web.NewTemplates(),
	}

	// Handle static files
	var staticFiles = http.FS(web.StaticFiles)
	fs := http.FileServer(staticFiles)
	gs.serveMux.Handle("/static/", http.StripPrefix("/static/", fs))

	// Serve all other requests, including websocket
	gs.serveMux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		gs.renderer.Render(w, "home", nil)
	})

	return gs
}

func (gs *GameServer) RunServer() {
	for {
		select {
		case client := <-gs.register:
			gs.Logf("Client %s registered.", client.Name)
			gs.clients[client.Name] = client

		case client := <-gs.unregister:
			gs.Logf("Client %s unregistered.", client.Name)
			delete(gs.clients, client.Name)

		case message := <-gs.broadcast:
			for _, client := range gs.clients {
				client.send <- message
			}
		}
	}
}

func (gs *GameServer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	gs.serveMux.ServeHTTP(w, r)
}

type returnData struct {
	name    string
	players []string
	room    string
	error   string
}

// Handle /createRoom endpoint
func (gs *GameServer) ServeCreateRoom(w http.ResponseWriter, r *http.Request) {
	playerName := r.URL.Query().Get("playerName")
	roomName := r.URL.Query().Get("roomName")

	// Input formats invalid error
	if len(playerName) < 1 || len(roomName) < 1 {
		w.Header().Set("HX-Retarget", "room-container")
		w.WriteHeader(http.StatusUnprocessableEntity)
		data := returnData{
			error: "Player name or Room name is invalid",
		}
		gs.renderer.Render(w, "chat", data)
		return
	}

	// Player name already exists error
	if _, exists := gs.clients[playerName]; exists {
		w.Header().Set("HX-Retarget", "room-container")
		w.WriteHeader(http.StatusUnprocessableEntity)
		data := returnData{
			error: "Player name exists. Pls choose a new one!",
		}
		gs.renderer.Render(w, "chat", data)
		return
	}

	// Room name alreay exists error
	if _, exists := gs.rooms[roomName]; exists {
		w.Header().Set("HX-Retarget", "room-container")
		w.WriteHeader(http.StatusUnprocessableEntity)
		data := returnData{
			error: "Room name exists. Pls choose a new one!",
		}
		gs.renderer.Render(w, "chat", data)
		return
	}

	data := returnData{
		name: playerName,
		players: []string{
			playerName,
		},
		room: roomName,
	}

	gs.renderer.Render(w, "chat", data)
}

// Handle /joinRoom endpoint
func (gs *GameServer) ServeJoinRoom(w http.ResponseWriter, r *http.Request) {
	playerName := r.URL.Query().Get("playerName")
	roomName := r.URL.Query().Get("roomName")

	// Input formats invalid error
	if len(playerName) < 1 || len(roomName) < 1 {
		w.Header().Set("HX-Retarget", "room-container")
		w.WriteHeader(http.StatusUnprocessableEntity)
		data := returnData{
			error: "Player name or Room name is invalid",
		}
		gs.renderer.Render(w, "chat", data)
		return
	}

	// Player name already exists error
	if _, exists := gs.clients[playerName]; exists {
		w.Header().Set("HX-Retarget", "room-container")
		w.WriteHeader(http.StatusUnprocessableEntity)
		data := returnData{
			error: "Player name exists. Pls choose a new one!",
		}
		gs.renderer.Render(w, "chat", data)
		return
	}

	// Room name doesn't exist error
	if _, exists := gs.rooms[roomName]; !exists {
		w.Header().Set("HX-Retarget", "room-container")
		w.WriteHeader(http.StatusUnprocessableEntity)
		data := returnData{
			error: "Room name doesn't exist.",
		}
		gs.renderer.Render(w, "chat", data)
		return
	}

	var players []string
	for player := range gs.rooms[roomName].clients {
		players = append(players, player)
	}
	data := returnData{
		name:    playerName,
		players: players,
		room:    roomName,
	}

	gs.renderer.Render(w, "chat", data)
}

// Handle websocket requests.
// Client should call ServeWebsocket if ServeCreateRoom or ServeJoinRoom is successful.
func (gs *GameServer) ServeWebsocket(w http.ResponseWriter, r *http.Request) {
	conn, err := websocket.Accept(w, r, nil)
	if err != nil {
		gs.Logf("%v\n", err)
		return
	}

	playerName := r.URL.Query().Get("player")
	roomName := r.URL.Query().Get("room")

	// Register new client
	client := NewClient(conn, playerName, gs)
	gs.register <- client

	readErr := client.ReadPump(r.Context())
	writeErr := client.WritePump(r.Context())

	select {
	case err := <-readErr:
		gs.Logf("%v\n", err)
	case err := <-writeErr:
		gs.Logf("%v\n", err)
	}

	// Register the new client in the correct room
	room, exists := gs.rooms[roomName]
	if !exists {
		room = gs.createRoom(client, roomName)
	}
	client.joinRoom(room, client)
}

func (gs *GameServer) createRoom(client *Client, roomName string) *Room {
	room := NewRoom(roomCapacity, roomName)
	room.clients[client.Name] = client
	go room.RunRoom()
	gs.rooms[room.Name] = room

	return room
}
