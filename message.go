package main

import (
	"log"
	"time"

	"github.com/gorilla/websocket"
)

// Message is ...
type Message struct {
	room string
	// message string
	message []byte
}

func broadcast(h *Hub, ch chan *Message) { // process data from redis pub sub
	for {
		msg := <-ch
		room := msg.room

		// send it out to every client in the room that is currently connected
		h.mtx.Lock()
		rooms := h.rooms[room]
		h.mtx.Unlock()

		for client := range rooms {
			client.ws.SetWriteDeadline(time.Now().Add(time.Millisecond * 150))
			err := client.ws.WriteMessage(websocket.TextMessage, []byte(msg.message))
			if err != nil {
				log.Println(err)
				h.unregister <- client
			}
		}
	}
}
