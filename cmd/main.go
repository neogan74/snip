package main

import (
	"log"
	"net/http"
)

func one(w http.ResponseWriter, r *http.Request) {
	// Check if the current request URL path exactly matches "/". If it doesn't, use
	// the http.NotFound() function to send a 404 response to the client.
	// Importantly, we then return from the handler. If we don't return the handler
	// would keep executing and also write the "Hello from SnippetBox" message.
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}

	w.Write([]byte("Hello. I'm Leo"))
}

func snippetView(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("Snippet View"))
}

func snippetCreate(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("Create Snippet"))
}

func main() {
	mux := http.NewServeMux()
	mux.HandleFunc("/", one)
	mux.HandleFunc("/snippet/view", snippetView)
	mux.HandleFunc("/snippet/create", snippetCreate)

	log.Println("Listening on :4000")
	err := http.ListenAndServe(":4000", mux)
	log.Fatal(err)
}
