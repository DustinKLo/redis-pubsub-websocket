package main

import (
	"fmt"
	"log"
	"sort"
	"time"

	"github.com/gomodule/redigo/redis"
	"github.com/gorilla/websocket"
)

const (
	writeWait = time.Millisecond * 100
)

// Hub is ...
type Hub struct {
	register    chan *Client
	unregister  chan *Client
	broadcast   chan *Message
	rooms       map[string]map[*Client]bool
	redisClient *redis.Pool
	psc         *redis.PubSubConn
}

// Message is ...
type Message struct {
	room    string
	message []byte
}

func newHub(psc *redis.PubSubConn) *Hub {
	return &Hub{
		register:   make(chan *Client),
		unregister: make(chan *Client),
		broadcast:  make(chan *Message),
		rooms:      make(map[string]map[*Client]bool),
		psc:        psc,
	}
}

func (h *Hub) redisListener() { //(ch chan *Message) {
	for {
		defer h.psc.Close()
		switch v := h.psc.Receive().(type) {
		case redis.Message:
			if debug == true {
				log.Println(string(v.Data))
			}
			h.broadcast <- &Message{v.Channel, v.Data}
		case redis.Subscription: // https://godoc.org/github.com/garyburd/redigo/redis#Subscription
			log.Println(v)
		case error:
			log.Printf("redis pubsub receive err: %v\n", v)
			panic("Redis connection broke")
		default:
			log.Println("something else happened")
		}
	}
}

func (h *Hub) printRoomsSize() {
	fmt.Println("####################################")
	keys := make([]string, 0)
	for k := range h.rooms {
		keys = append(keys, k)
	}
	if len(keys) > 0 {
		sort.Strings(keys)
		for _, k := range keys {
			fmt.Println(k, len(h.rooms[k]))
		}
	} else {
		fmt.Println("NO MORE CLIENTS IN H.ROOMS")
	}
	fmt.Println("####################################")
}

func (h *Hub) registerUser(c *Client) {
	for _, room := range c.rooms {
		if h.rooms[room] == nil {
			h.rooms[room] = make(map[*Client]bool)
			h.psc.Subscribe(room)
		}
		h.rooms[room][c] = true
	}
	log.Println("client registered", c.conn.RemoteAddr())
	if debug == true {
		h.printRoomsSize()
	}
}

func (h *Hub) unregisterUser(c *Client) {
	for _, room := range c.rooms {
		delete(h.rooms[room], c)
		if h.rooms[room] != nil && len(h.rooms[room]) == 0 {
			delete(h.rooms, room)
			h.psc.Unsubscribe(room)
		}
		if debug == true {
			h.printRoomsSize()
		}
	}
	log.Println("client UN-registered", c.conn.RemoteAddr())
	c.conn.Close()
}

func (h *Hub) run() { //(ch chan *Message) {
	for {
		select {
		case c := <-h.register:
			h.registerUser(c)
		case c := <-h.unregister:
			h.unregisterUser(c)
		case msg := <-h.broadcast:
			for c := range h.rooms[msg.room] {
				c.conn.SetWriteDeadline(time.Now().Add(writeWait))

				err := c.conn.WriteMessage(websocket.TextMessage, []byte(msg.message))
				if err != nil {
					log.Println("Sent message err: ", err)
					h.unregisterUser(c)
				}
			}
		}
	}
}
