package main

import (
	"log"

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

func (h *Hub) run(rHub *RedisHub, ch chan *Message) {
	for {
		select {
		case client := <-h.register:
			log.Println("registered client: ", client)
			for _, room := range client.rooms {
				if h.rooms[room] == nil {
					h.rooms[room] = make(map[*Client]bool)
					go rHub.subClient(room, ch)
				}
				h.rooms[room][client] = true
			}
		case client := <-h.unregister:
			log.Println("un-registered client: ", client)
			for _, room := range client.rooms {
				delete(h.rooms[room], client)
				if len(h.rooms[room]) == 0 {
					delete(h.rooms, room)
					rHub.channels[room].Unsubscribe()
				}
			}
			client.ws.Close()
		}
	}
}
