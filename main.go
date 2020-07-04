package main

import (
	"log"
	"net/http"
	"strings"

	"github.com/gomodule/redigo/redis"
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
	wsClient := &Client{ws, rooms}
	h.register <- wsClient

	_, _, readErr := ws.ReadMessage() // detecting when client closes
	if readErr != nil {
		h.unregister <- wsClient
	}
}

func main() {
	c := redisConn("127.0.0.1:6379")
	psc := redis.PubSubConn{Conn: c}
	psc.PSubscribe("*")

	msgCh := make(chan Message) // REDIS PUB SUB CHHANNEL

	hub := createHub()
	go hub.run()
	go subClient(psc, msgCh)
	go broadcastMsg(hub, msgCh) // process data from redis pub sub

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
