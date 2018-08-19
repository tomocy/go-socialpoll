package main

import (
	"flag"
	"log"
	"net/http"
)

func main() {
	addr := flag.String("addr", ":8081", "the address of the web")
	flag.Parse()
	startListeningAndServing(*addr)
}

func startListeningAndServing(addr string) {
	mux := http.NewServeMux()
	setRouting(mux)
	log.Println("listening: ", addr)
	http.ListenAndServe(addr, mux)
}

func setRouting(mux *http.ServeMux) {
	mux.Handle("/", http.StripPrefix("/", http.FileServer(http.Dir("public"))))
}
