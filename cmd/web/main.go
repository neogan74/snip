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
	sessionManager *scs.SessionManager
	users          *models.UserModel
}

func main() {
	var cfg Config
	flag.StringVar(&cfg.addr, "addr", "localhost:4001", "http service address")
	flag.Parse()

	errorLog := log.New(os.Stdout, "ERROR\t", log.Ldate|log.Ltime|log.Lshortfile)
	infoLog := log.New(os.Stdout, "INFO\t", log.Ldate|log.Ltime)

	// Database connection
	// Ensure SNIP_DB_DSN environment variable is set
	dsn := os.Getenv("SNIP_DB_DSN")
	if dsn == "" {
		errorLog.Fatal("SNIP_DB_DSN environment variable not set")
	}
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		errorLog.Fatal(err)
	}
	defer db.Close()

	if err := db.Ping(); err != nil {
		errorLog.Fatal(err)
	}
	infoLog.Println("Connected to the database successfully")

	templateCache, err := newTemplateCache()

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
		snippets:       &models.SnippetModel{DB: db},
		users:          &models.UserModel{DB: db},
		templateCache:  templateCache,
		formDecoder:    formDecoder,
		sessionManager: sessionManager,
	}

	tlsConfig := &tls.Config{
		CurvePreferences: []tls.CurveID{tls.X25519, tls.CurveP256},
	}
	srv := &http.Server{
		Addr:         cfg.addr,
		Handler:      app.routes(),
		ErrorLog:     errorLog,
		TLSConfig:    tlsConfig,
		IdleTimeout:  time.Minute,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
	}

	infoLog.Printf("Listening on https://%s\n", cfg.addr)
	err = srv.ListenAndServeTLS("./tls/cert.pem", "./tls/key.pem")
	errorLog.Fatal(err)
}
