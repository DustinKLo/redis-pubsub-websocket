package main

import (
	"log"

	"github.com/gomodule/redigo/redis"
	"github.com/gorilla/websocket"
)

// Client is ...
type Client struct {
	ws    *websocket.Conn
	rooms []string
}

// Hub is ...
type Hub struct {
	register   chan *Client
	unregister chan *Client
	rooms      map[string]map[*Client]bool
}

func createHub() *Hub {
	return &Hub{
		register:   make(chan *Client),
		unregister: make(chan *Client),
		rooms:      make(map[string]map[*Client]bool),
	}
}

func (h *Hub) run(psc *redis.PubSubConn) {
	for {
		select {
		case client := <-h.register:
			log.Println("registered client: ", client)
			for _, room := range client.rooms {
				if h.rooms[room] == nil {
					h.rooms[room] = make(map[*Client]bool)
					log.Println("Subscribing to room", room)
					psc.Subscribe(room)
				}
				h.rooms[room][client] = true
			}
		case client := <-h.unregister:
			log.Println("un-registered client: ", client)
			for _, room := range client.rooms {
				delete(h.rooms[room], client)
				if len(h.rooms[room]) == 0 {
					delete(h.rooms, room)
					log.Println("Un-Subscribing to room", room)
					psc.Unsubscribe(room)
				}
			}
			client.ws.Close()
		}
	}
}
