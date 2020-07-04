package main

import "github.com/gorilla/websocket"

type Message struct {
	room    string
	message string
}

func broadcastMsg(h *Hub, ch chan *Message) { // process data from redis pub sub
	for {
		msg := <-ch
		room := msg.room
		// send it out to every client in the room that is currently connected
		for client := range h.rooms[room] {
			err := client.ws.WriteMessage(websocket.TextMessage, []byte(msg.message))
			if err != nil {
				h.unregister <- client
			}
		}
	}
}
