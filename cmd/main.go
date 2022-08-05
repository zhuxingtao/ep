package main

import (
	"ep"
	"net/http"
)

func init() {
	http.HandleFunc("/ws", HandleWs)
}

func main() {
	hub = ep.NewHub()
	hub.Start()
	http.ListenAndServe(":8080", nil)
}
