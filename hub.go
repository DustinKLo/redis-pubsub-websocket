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

func (h *Hub) run(r *RedisHub, ch chan *Message) {
	count := 0
	for {
		select {
		case client := <-h.register:
			h.mtx.Lock()
			for _, room := range client.rooms {
				if h.rooms[room] == nil {
					h.rooms[room] = make(map[*Client]bool)
					r.subscribe <- room
				}
				h.rooms[room][client] = true
			}
			count++
			h.mtx.Unlock()
		case client := <-h.unregister:
			h.mtx.Lock()
			for _, room := range client.rooms {
				delete(h.rooms[room], client)
				if h.rooms[room] != nil && len(h.rooms[room]) == 0 {
					delete(h.rooms, room)
				}
			}
			count--
			h.mtx.Unlock()
			client.ws.Close()
		}
		log.Println(count, "clients registered")
	}
}
