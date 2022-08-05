package main

import (
	"ep"
	"github.com/gorilla/websocket"
	"log"
	"net/http"
)

var (
	hub      *ep.Hub
	upGrader = websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool {
			return true
		},
	}
)

func HandleWs(w http.ResponseWriter, r *http.Request) {
	p := r.URL.Query()
	uid := p.Get("uid")
	if uid == "" {
		uid = "1234"
	}
	c, err := upGrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println("/ws upgrade fail", err)
		return
	}
	clientID := ep.StrUUID()
	cli := ep.NewClient(c, clientID)
	cli.UID = uid
	err = hub.Add(cli)
	if err != nil {
		log.Println("hub add cli fail", cli, err)
	}
	cli.Start()
}
