package main

import (
	"flag"
	"log"
	"net/http"
)

func main() {
	addr := flag.String("addr", ":8081", "the address of the web")
	flag.Parse()

	mux := http.NewServeMux()
	mux.Handle("/", http.StripPrefix("/", http.FileServer(http.Dir("public"))))
	log.Println("listening: ", *addr)
	http.ListenAndServe(*addr, mux)
}
