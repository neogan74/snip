package main

import (
	"flag"
	"log"
	"net/http"
	"os"
)

type Config struct {
	addr string
}

type App struct {
	errorLog *log.Logger
	infoLog  *log.Logger
}

func main() {
	var cfg Config
	flag.StringVar(&cfg.addr, "addr", "localhost:4000", "http service address")
	flag.Parse()

	errorLog := log.New(os.Stdout, "ERROR\t", log.Ldate|log.Ltime|log.Lshortfile)
	infoLog := log.New(os.Stdout, "INFO\t", log.Ldate|log.Ltime)
	// DI
	app := &App{
		errorLog: errorLog,
		infoLog:  infoLog,
	}
	srv := &http.Server{
		Addr:     cfg.addr,
		Handler:  app.routes(),
		ErrorLog: errorLog,
	}

	infoLog.Printf("Listening on http://%s\n", cfg.addr)
	err := srv.ListenAndServe()
	errorLog.Fatal(err)
}
