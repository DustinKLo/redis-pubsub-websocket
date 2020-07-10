package main

import (
	"log"
	"sync"

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
	mtx        sync.Mutex
}

func newHub() *Hub {
	return &Hub{
		register:   make(chan *Client),
		unregister: make(chan *Client),
		rooms:      make(map[string]map[*Client]bool),
		mtx:        sync.Mutex{},
	}
}

func (h *Hub) run(rHub *RedisHub, ch chan *Message) {
	count := 0
	for {
		select {
		case client := <-h.register:
			// log.Println("registered client: ", client)
			h.mtx.Lock()
			for _, room := range client.rooms {
				if h.rooms[room] == nil {
					h.rooms[room] = make(map[*Client]bool)
					go rHub.subClient(room, ch)
				}
				h.rooms[room][client] = true
			}
			h.mtx.Unlock()
			count++
		case client := <-h.unregister:
			// log.Println("un-registered client: ", client)
			h.mtx.Lock()
			for _, room := range client.rooms {
				delete(h.rooms[room], client)
				if len(h.rooms[room]) == 0 {
					delete(h.rooms, room)
					rHub.channels[room].Unsubscribe()
				}
			}
			h.mtx.Unlock()
			client.ws.Close()
			count--
		}
		log.Println(count, "clients registered")
	}
}
