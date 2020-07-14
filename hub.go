package main

import (
	"log"
)

// Hub is ...
type Hub struct {
	register   chan *Client
	unregister chan *Client
	rooms      map[string]map[*Client]bool
}

// Message is ...
type Message struct {
	room    string
	message []byte
}

func newHub() *Hub {
	return &Hub{
		register:   make(chan *Client),
		unregister: make(chan *Client),
		rooms:      make(map[string]map[*Client]bool),
	}
}

func (h *Hub) run(r *RedisHub, ch chan *Message) {
	for {
		select {
		case c := <-h.register:
			for _, room := range c.rooms {
				if h.rooms[room] == nil {
					h.rooms[room] = make(map[*Client]bool)
					r.subscribe <- room
				}
				h.rooms[room][c] = true
			}
			log.Println("client registered", c.conn.RemoteAddr())

		case c := <-h.unregister:
			c.closeOnce.Do(func() {
				close(c.send)
			})
			for _, room := range c.rooms {
				delete(h.rooms[room], c)
				if h.rooms[room] != nil && len(h.rooms[room]) == 0 {
					delete(h.rooms, room)
					r.unsubscribe <- room
				}
			}
			log.Println("client UN-registered", c.conn.RemoteAddr())

		case msg := <-ch:
			// log.Println(string(msg.message))
			for c := range h.rooms[msg.room] {
				c.send <- msg.message
			}
		}
	}
}
