package main

import (
	"flag"
	"log"
	"net/http"
	"time"

	"github.com/Livingpool/tim-tris/server"
)

var addr = flag.String("addr", ":42069", "http server addr")

func main() {
	flag.Parse()

	server := &http.Server{
		Addr:         *addr,
		Handler:      server.NewGameServer(),
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 30 * time.Second,
	}

	log.Println("Server listening on port", *addr)
	log.Fatal(server.ListenAndServe())
}
