package main

import (
	"flag"
	"log"
	"net/http"
)

func main() {
	// Define a new command-line flag for address
	addr := flag.String("addr", "localhost:4000", "http service address")
	flag.Parse()
	mux := http.NewServeMux()

	fileServer := http.FileServer(http.Dir("./ui/static"))

	mux.Handle("/static/", http.StripPrefix("/static/", fileServer))
	mux.HandleFunc("/", home)
	mux.HandleFunc("/snippet/view", snippetView)
	mux.HandleFunc("/snippet/create", snippetCreate)

	log.Printf("Listening on http://%s\n", *addr)
	err := http.ListenAndServe(*addr, mux)
	log.Fatal(err)
}
