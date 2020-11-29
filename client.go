package main

import (
	"log"
	"net/http"
	"strings"

	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
)

// Client is ...
type Client struct {
	conn  *websocket.Conn
	hub   *Hub
	rooms []string
}

func newClient(conn *websocket.Conn, h *Hub, rooms []string) *Client {
	return &Client{
		conn:  conn,
		hub:   h,
		rooms: rooms,
	}
}

func handleWS(h *Hub, w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	ws, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Fatal("handleWS err: ", err)
		return
	}
	rooms := strings.Split(vars["rooms"], ",")
	c := newClient(ws, h, rooms)

	h.register <- c

	_, _, err = c.conn.ReadMessage() // detecting when client closes
	if err != nil {
		log.Println("Client closed: ", c.conn.RemoteAddr(), err.Error())
		c.hub.unregister <- c
		return
	}
}
