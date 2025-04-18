package main

import (
	"crypto/tls"
	"database/sql"
	"flag"
	"html/template"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/alexedwards/scs/mysqlstore"
	"github.com/alexedwards/scs/v2"
	"github.com/go-playground/form/v4"
	"github.com/neogan74/snip/internal/models"

	_ "github.com/go-sql-driver/mysql"
)

type Config struct {
	addr string
}

type App struct {
	errorLog       *log.Logger
	infoLog        *log.Logger
	snippets       *models.SnippetModel
	templateCache  map[string]*template.Template
	formDecoder    *form.Decoder
	seesionManager *scs.SessionManager
}

func main() {
	var cfg Config
	flag.StringVar(&cfg.addr, "addr", "localhost:4001", "http service address")
	flag.Parse()

	errorLog := log.New(os.Stdout, "ERROR\t", log.Ldate|log.Ltime|log.Lshortfile)
	infoLog := log.New(os.Stdout, "INFO\t", log.Ldate|log.Ltime)

	db, err := sql.Open("mysql", "snipuser:snippassword@/snipdb?parseTime=true")
	if err != nil {
		errorLog.Fatal(err)
	}
	defer db.Close()
	infoLog.Println(db.Ping())

	templateCache, err := newTempalteCache()

	if err != nil {
		errorLog.Fatal(err)
	}

	formDecoder := form.NewDecoder()

	// Session
	sessionManager := scs.New()
	sessionManager.Store = mysqlstore.New(db)
	sessionManager.Lifetime = 24 * time.Hour
	// DI
	app := &App{
		errorLog:       errorLog,
		infoLog:        infoLog,
		snippets:       &models.SnippetModel{db},
		templateCache:  templateCache,
		formDecoder:    formDecoder,
		seesionManager: sessionManager,
	}

	tlsConfig := &tls.Config{
		CurvePreferences: []tls.CurveID{tls.X25519, tls.CurveP256},
	}
	srv := &http.Server{
		Addr:      cfg.addr,
		Handler:   app.routes(),
		ErrorLog:  errorLog,
		TLSConfig: tlsConfig,
	}

	infoLog.Printf("Listening on https://%s\n", cfg.addr)
	err = srv.ListenAndServeTLS("./tls/cert.pem", "./tls/key.pem")
	errorLog.Fatal(err)
}
