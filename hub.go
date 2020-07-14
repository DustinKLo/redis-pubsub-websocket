package main

import (
	"log"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

const writeWait = time.Millisecond * 100

// Client is ...
type Client struct {
	conn        *websocket.Conn
	hub         *Hub
	rooms       []string
	send        chan []byte
	closeClient sync.Once
}

func newClient(conn *websocket.Conn, h *Hub, rooms []string) *Client {
	return &Client{
		conn:        conn,
		hub:         h,
		rooms:       rooms,
		send:        make(chan []byte),
		closeClient: sync.Once{},
	}
}

func (c *Client) readPump() {
	defer func() {
		c.conn.Close()
		c.hub.unregister <- c
	}()

	c.hub.register <- c

	_, _, rErr := c.conn.ReadMessage() // detecting when client closes
	if rErr != nil {
		return
	}
}

func (c *Client) writePump() { // writing messages to the websocket client
	defer c.conn.Close()

	for {
		select {
		case msg, ok := <-c.send:
			if !ok {
				return
			}
			writeWait := time.Now().Add(writeWait)
			c.conn.SetWriteDeadline(writeWait)

			err := c.conn.WriteMessage(websocket.TextMessage, []byte(msg))
			if err != nil {
				log.Println(err)
				c.conn.Close()
			}
		}
	}
}

// Hub is ...
type Hub struct {
	register   chan *Client
	unregister chan *Client
	rooms      map[string]map[*Client]bool
	mtx        sync.Mutex
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
		mtx:        sync.Mutex{},
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
			c.closeClient.Do(func() {
				close(c.send)
			})
			for _, room := range c.rooms {
				delete(h.rooms[room], c)
				if h.rooms[room] != nil && len(h.rooms[room]) == 0 {
					delete(h.rooms, room)
					r.unsubscribe <- room
				}
			}
			log.Println("client un-registered", c.conn.RemoteAddr())

		case msg := <-ch:
			// log.Println(string(msg.message))
			for c := range h.rooms[msg.room] {
				c.send <- msg.message
			}
		}
	}
}
