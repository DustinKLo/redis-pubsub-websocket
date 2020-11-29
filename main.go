package main

import (
	"flag"
	"log"
	"net/http"
	"os"

	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true // CORS?
	},
}

func renderHomePage(w http.ResponseWriter, r *http.Request) {
	path, _ := os.Getwd()
	http.ServeFile(w, r, path+"/templates/index.html")
}

var (
	debug bool
)

func main() {
	var (
		redisHost string
	)

	flag.StringVar(&redisHost, "redis", "redis://127.0.0.1:6379", "redis endpoint (default: redis://127.0.0.1:6379)")
	flag.BoolVar(&debug, "debug", false, "debug mode, stdout results")
	flag.Parse()

	redisPool := newRedisClient(redisHost)
	psc := newPubsubClient(redisPool)

	hub := newHub(psc)
	go hub.redisListener()
	go hub.run()

	r := mux.NewRouter()
	r.HandleFunc("/", renderHomePage)
	r.HandleFunc("/ws/{rooms}", func(w http.ResponseWriter, r *http.Request) {
		handleWS(hub, w, r)
	})

	log.Println("http server started on :8000") // starting server
	err := http.ListenAndServe(":8000", r)
	if err != nil {
		log.Fatal("ListAndServe: ", err)
	}
}
