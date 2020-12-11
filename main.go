package main

import (
	"flag"
	"fmt"
	"net/http"

	"github.com/sirupsen/logrus"

	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true // CORS?
	},
}

var (
	debug bool
)

func main() {
	var (
		port      int
		redisHost string
	)

	flag.IntVar(&port, "port", 8000, "server port number")
	flag.StringVar(&redisHost, "redis", "redis://127.0.0.1:6379", "redis endpoint")
	flag.BoolVar(&debug, "debug", false, "debug mode, stdout results")
	flag.Parse()

	if debug == true {
		logger.SetLevel(logrus.DebugLevel)
	}

	redisPool := newRedisPool(redisHost)
	psc := newPubsubClient(redisPool)

	hub := newHub(psc)
	go hub.redisListener()
	go hub.run()

	r := mux.NewRouter()
	r.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "templates/index.html")
	})
	r.HandleFunc("/ws/{rooms}", func(w http.ResponseWriter, r *http.Request) {
		handleWS(hub, w, r)
	})

	logger.Infoln(fmt.Sprintf("http server started on :%d", port)) // starting server
	if err := http.ListenAndServe(fmt.Sprintf(":%d", port), r); err != nil {
		logger.Fatalln("ListAndServe: ", err)
	}
}
