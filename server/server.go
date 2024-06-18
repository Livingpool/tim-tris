package server

import (
	"net/http"

	"github.com/Livingpool/tim-tris/client"
)

type GameServer struct {

	// serveMux routes the various endpoints to the appropriate handler.
	serveMux http.ServeMux

	rooms map[string]client.Room
}
