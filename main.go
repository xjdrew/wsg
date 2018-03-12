package main

import (
	"flag"
	"log"
	"net/http"
)

var addr = flag.String("addr", ":8080", "websocket service address")
var server = flag.String("server", "127.0.0.1:8081", "backend server address")

func main() {
	flag.Parse()

	server := newServer(*server)
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		serveWs(server, w, r)
	})
	err := http.ListenAndServe(*addr, nil)
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}
