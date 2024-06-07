package router

import "net/http"

func Init() *http.ServeMux {
	router := http.NewServeMux()

	router.HandleFunc("/", handddler.HomePage)
	router.HandleFunc("/")
	router.Handle("/", room.NewRoom())
}
