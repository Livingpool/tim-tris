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

	gameServer := server.NewGameServer()
	go gameServer.RunServer()

	s := &http.Server{
		Addr:         *addr,
		Handler:      gameServer,
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 30 * time.Second,
	}

	log.Println("Server listening on port", *addr)
	log.Fatal(s.ListenAndServe())
}
