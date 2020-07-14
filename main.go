package main

import (
	"log"
	"net/http"
	"strings"

	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true // CORS?
	},
}

func handleWS(h *Hub, w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	ws, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Fatal(err)
	}
	rooms := strings.Split(vars["rooms"], ",")
	c := newClient(ws, h, rooms)
	go c.readPump()
	go c.writePump()
}

func main() {
	msgCh := make(chan *Message) // go channel to hold all messages to broadcast

	rPool := newRedisPool("redis://127.0.0.1:6379")
	rConn := rPool.Get()
	rHub := newRedisHub(&rConn)
	go rHub.subscribeHandler()
	go rHub.subClient(msgCh)

	hub := newHub()
	go hub.run(rHub, msgCh)

	r := mux.NewRouter()
	r.HandleFunc("/ws/{rooms}", func(w http.ResponseWriter, r *http.Request) {
		handleWS(hub, w, r)
	})

	log.Println("http server started on :8000") // starting server
	err := http.ListenAndServe(":8000", r)
	if err != nil {
		log.Fatal("ListAndServe: ", err)
	}
}
