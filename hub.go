package main

import (
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
	register   chan *Client
	unregister chan *Client
	broadcast  chan *Message
	rooms      map[string]map[*Client]bool
	psc        *redis.PubSubConn
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
				logger.Debugln(string(v.Data))
			}
			h.broadcast <- &Message{v.Channel, v.Data}
		case redis.Subscription: // https://godoc.org/github.com/garyburd/redigo/redis#Subscription
			logger.Infof("redis: %s to %s\n", v.Kind, v.Channel)
		case error:
			logger.Panicf("redis err: %v\n", v)
			panic("Redis connection broke")
		default:
			logger.Warningln("redis: something else happened")
		}
	}
}

func (h *Hub) printRoomsSize() {
	logger.Debugln("####################################")
	keys := make([]string, 0)
	for k := range h.rooms {
		keys = append(keys, k)
	}
	if len(keys) > 0 {
		sort.Strings(keys)
		for _, k := range keys {
			logger.Debugln(k, len(h.rooms[k]))
		}
	} else {
		logger.Debugln("NO MORE CLIENTS IN H.ROOMS")
	}
	logger.Debugln("####################################")
}

func (h *Hub) registerUser(c *Client) {
	for _, room := range c.rooms {
		if h.rooms[room] == nil {
			h.rooms[room] = make(map[*Client]bool)
			h.psc.Subscribe(room)
		}
		h.rooms[room][c] = true
	}
	logger.Infoln("client registered", c.conn.RemoteAddr())
	if debug == true {
		h.printRoomsSize()
	}
}

func (h *Hub) unregisterUser(c *Client) {
	defer c.conn.Close()

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
	logger.Infoln("client UN-registered", c.conn.RemoteAddr())
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

				if err := c.conn.WriteMessage(websocket.TextMessage, []byte(msg.message)); err != nil {
					logger.Warningln("Sent message err: ", err)
					h.unregisterUser(c)
				}
			}
		}
	}
}
