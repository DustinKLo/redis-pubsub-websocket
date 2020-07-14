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
	conn      *websocket.Conn
	hub       *Hub
	rooms     []string
	send      chan []byte
	closeOnce sync.Once
}

func newClient(conn *websocket.Conn, h *Hub, rooms []string) *Client {
	return &Client{
		conn:      conn,
		hub:       h,
		rooms:     rooms,
		send:      make(chan []byte),
		closeOnce: sync.Once{},
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
