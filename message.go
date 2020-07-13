package main

import (
	"log"
	"time"

	"github.com/gorilla/websocket"
)

// Message is ...
type Message struct {
	room    string
	message []byte
}

func broadcast(h *Hub, ch chan *Message) { // process data from redis pub sub
	for {
		msg := <-ch

		// send it out to every client in the room that is currently connected
		h.mtx.Lock()
		for client := range h.rooms[msg.room] {
			readDeadline := time.Now().Add(time.Millisecond * 150)
			client.ws.SetWriteDeadline(readDeadline)
			err := client.ws.WriteMessage(websocket.TextMessage, []byte(msg.message))
			if err != nil {
				log.Println(err)
				// h.unregister <- client
			}
		}
		h.mtx.Unlock()
	}
}
