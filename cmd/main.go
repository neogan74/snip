package main

import (
	"log"
	"net/http"
)

func one(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("Hello. I'm Leo"))
}

func main() {
	mux := http.NewServeMux()
	mux.HandleFunc("/", one)
	log.Println("Listening on :4000")
	err := http.ListenAndServe(":4000", mux)
	log.Fatal(err)
}
