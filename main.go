package main

import (
	"flag"
	"net/http"
	"os"

	"github.com/sirupsen/logrus"

	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
)

func renderHomePage(w http.ResponseWriter, r *http.Request) {
	path, _ := os.Getwd()
	http.ServeFile(w, r, path+"/templates/index.html")
}

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
		redisHost string
	)

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
	r.HandleFunc("/", renderHomePage)
	r.HandleFunc("/ws/{rooms}", func(w http.ResponseWriter, r *http.Request) {
		handleWS(hub, w, r)
	})

	logger.Infoln("http server started on :8000") // starting server
	if err := http.ListenAndServe(":8000", r); err != nil {
		logger.Fatalln("ListAndServe: ", err)
	}
}
