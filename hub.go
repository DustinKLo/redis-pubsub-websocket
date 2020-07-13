package main

import (
	"log"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

// var writeDeadline = time.Now().Add(time.Millisecond * 150)
var writeDeadline = time.Now().Add(time.Second * 10)

// Client is ...
type Client struct {
	conn  *websocket.Conn
	hub   *Hub
	rooms []string
	send  chan []byte
}

func newClient(conn *websocket.Conn, h *Hub, rooms []string) *Client {
	return &Client{
		conn:  conn,
		hub:   h,
		rooms: rooms,
		send:  make(chan []byte),
	}
}

func (c *Client) writePump() {
	// writing messages to the websocket client
	for {
		select {
		case msg, ok := <-c.send:
			if !ok {
				// The hub closed the channel.
				c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}
			// log.Println(string(msg))
			// c.conn.SetWriteDeadline(writeDeadline)
			err := c.conn.WriteMessage(websocket.TextMessage, []byte(msg))
			if err != nil {
				log.Println(err)
			}
			// w, err := c.conn.NextWriter(websocket.TextMessage)
			// if err != nil {
			// 	log.Println(err)
			// }
			// w.Write(msg)
		}
	}
}

func (c *Client) readPump() {
	defer func() {
		c.conn.Close()
		close(c.send)
	}()

	c.hub.register <- c

	_, _, rErr := c.conn.ReadMessage() // detecting when client closes
	if rErr != nil {
		c.hub.unregister <- c
		return
	}
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
					r.unsubscribe <- room
				}
			}
			count--
			h.mtx.Unlock()
			client.conn.Close()
		}
		log.Println(count, "clients registered")
	}
}
