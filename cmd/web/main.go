package main

import (
	"flag"
	"log"
	"net/http"
	"os"
)

type Config struct {
	addr      string
	staticDir string
}

type App struct {
	errorLog *log.Logger
	infoLog  *log.Logger
}

func main() {
	var cfg Config
	flag.StringVar(&cfg.addr, "addr", "localhost:4000", "http service address")
	flag.StringVar(&cfg.staticDir, "static-dir", "./ui/static", "Path to static assets")
	flag.Parse()
	errorLog := log.New(os.Stdout, "ERROR\t", log.Ldate|log.Ltime|log.Lshortfile)
	infoLog := log.New(os.Stdout, "INFO\t", log.Ldate|log.Ltime)
	// DI
	app := &App{
		errorLog: errorLog,
		infoLog:  infoLog,
	}

	// Logging

	mux := http.NewServeMux()

	fileServer := http.FileServer(http.Dir(cfg.staticDir))
	mux.Handle("/static/", http.StripPrefix("/static/", fileServer))

	mux.HandleFunc("/", app.home)
	mux.HandleFunc("/snippet/view", app.snippetView)
	mux.HandleFunc("/snippet/create", app.snippetCreate)

	srv := &http.Server{
		Addr:     cfg.addr,
		Handler:  mux,
		ErrorLog: errorLog,
	}

	infoLog.Printf("Listening on http://%s\n", cfg.addr)
	err := srv.ListenAndServe()
	errorLog.Fatal(err)
}
