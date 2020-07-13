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

func handleWSConns(h *Hub, w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	ws, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Fatal(err)
	}
	rooms := strings.Split(vars["rooms"], ",")
	client := &Client{ws, rooms}
	h.register <- client

	_, _, rErr := ws.ReadMessage() // detecting when client closes
	if rErr != nil {
		h.unregister <- client
		ws.Close()
	}
}

func main() {
	msgCh := make(chan *Message) // go channel to hold all messages to broadcast

	redisPool := newRedisPool("redis://127.0.0.1:6379")
	redisConn := redisPool.Get()
	redisHub := newRedisHub(&redisConn)
	go redisHub.subscribeHandler()
	go redisHub.subClient(msgCh)

	hub := newHub()
	go hub.run(redisHub, msgCh)
	go broadcast(hub, msgCh) // process data from redis pub sub

	r := mux.NewRouter()
	r.HandleFunc("/ws/{rooms}", func(w http.ResponseWriter, r *http.Request) {
		handleWSConns(hub, w, r)
	})

	log.Println("http server started on :8000") // starting server
	err := http.ListenAndServe(":8000", r)
	if err != nil {
		log.Fatal("ListAndServe: ", err)
	}
}
