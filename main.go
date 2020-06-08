package main

import (
	"log"
	"net/http"

	"github.com/gomodule/redigo/redis"
	"github.com/gorilla/websocket"
)

var msgCh = make(chan string) // REDIS PUB SUB CHHANNEL
var clients = make(map[*websocket.Conn]bool)
var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true // CORS?
	},
}

func handleWSConns(w http.ResponseWriter, r *http.Request) {
	ws, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Fatal(err)
	} else {
		log.Println(ws.RemoteAddr())
	}
	clients[ws] = true
	log.Println(clients)

	_, _, readErr := ws.ReadMessage() // detecting when client closes
	if readErr != nil {
		log.Println("removing client", ws.RemoteAddr())
		delete(clients, ws)
		ws.Close()
	}
}

func broadcastMsg(ch chan string) { // process data from redis pub sub
	for {
		msg := <-ch
		log.Println("websocket clients:", clients)

		go func() { // not sure if i should wrap this in a separate go routine
			// send it out to every client that is currently connected
			for client := range clients {
				err := client.WriteMessage(websocket.TextMessage, []byte(msg))
				if err != nil {
					client.Close()
					delete(clients, client)
				}
			}
		}()
	}
}

func subClient(psc redis.PubSubConn) {
	for {
		defer psc.Close()
		switch v := psc.Receive().(type) {
		case redis.Message:
			log.Printf("[%s] %s\n", v.Channel, v.Data)
			msgCh <- string(v.Data) // maybe store it as bytes
		case redis.Subscription:
			log.Printf("Subscribed to redis pub sub channel %s: %s %d\n", v.Channel, v.Kind, v.Count)
		case error:
			log.Printf("redis pubsub receive err: %v\n", v)
			panic("Redis Sub connection broke")
		default:
			log.Println("something else happened")
		}
	}
}

func redisConn(host string) redis.Conn {
	c, err := redis.Dial("tcp", host)
	if err != nil {
		panic(err)
	}
	return c
}

func main() {
	http.HandleFunc("/ws", handleWSConns)

	pubsubChan := "test"
	c := redisConn("127.0.0.1:6379")
	psc := redis.PubSubConn{Conn: c}
	psc.Subscribe(pubsubChan)

	go subClient(psc)
	go broadcastMsg(msgCh) // process data from redis pub sub

	log.Println("http server started on :8000")
	err := http.ListenAndServe(":8000", nil)
	if err != nil {
		log.Fatal("ListAndServe: ", err)
	}
}
