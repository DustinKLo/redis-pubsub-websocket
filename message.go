package main

import (
	"log"
	"time"

	"github.com/gorilla/websocket"
)

// Message is ...
type Message struct {
	room    string
	message string
}

func broadcast(h *Hub, ch chan *Message) { // process data from redis pub sub
	for {
		msg := <-ch
		room := msg.room
		// send it out to every client in the room that is currently connected
		for client := range h.rooms[room] {
			client.ws.SetWriteDeadline(time.Now().Add(time.Millisecond * 150))
			err := client.ws.WriteMessage(websocket.TextMessage, []byte(msg.message))
			if err != nil {
				log.Println(err)
				h.unregister <- client
			}
		}
	}
}
