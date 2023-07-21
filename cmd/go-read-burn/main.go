package main

import (
	"context"
	"embed"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"path"
	"text/template"
	"time"

	"github.com/boltdb/bolt"
	"github.com/gorilla/mux"
	"github.com/kelseyhightower/envconfig"
)

//go:embed all:views/*
var views embed.FS

//go:embed static/*
var static embed.FS

var (
	db        *bolt.DB
	templates *template.Template
	version   = "0.0.0-development"
	commit    = "none"
	date      = "unknown"
)

type Config struct {
	DBPath     string `default:"db/secrets.db" split_words:"true"`
	ListenPort string `default:"80" split_words:"true"`
	ListenHost string `default:"0.0.0.0" split_words:"true"`
}

// Main entry point for the app.
func main() {
	log.Printf("Version %s - Commit: %s, Build Date: %s", version, commit, date)

	// Read config
	var config Config
	var err error
	if err = envconfig.Process("GRB", &config); err != nil {
		log.Println(err)
	}

	// Open the DB
	if err = createDBDir(config.DBPath); err != nil {
		log.Fatalf("failed to create database directory: %v", err)
	}
	db, err = bolt.Open(config.DBPath, 0644, nil)
	if err != nil {
		log.Println(err)
	}
	defer db.Close()

	r := mux.NewRouter()
	r.HandleFunc("/", IndexHandler)
	r.HandleFunc("/create", CreateHandler).Methods("POST")
	r.HandleFunc("/get/{key}", SecretHandler)
	s := http.StripPrefix("/static/", http.FileServer(http.FS(static)))
	r.PathPrefix("/static/").Handler(s)
	http.Handle("/", r)

	templates = template.Must(template.ParseFS(views, "views/*.html"))

	srv := &http.Server{
		Handler:      r,
		Addr:         fmt.Sprintf("%s:%s", config.ListenHost, config.ListenPort),
		WriteTimeout: 15 * time.Second,
		ReadTimeout:  15 * time.Second,
	}

	// Run our server in a goroutine so that it doesn't block.
	go func() {
		log.Printf("Started server listing on %s:%s", config.ListenHost, config.ListenPort)
		if err := srv.ListenAndServe(); err != nil {
			log.Println(err)
		}
	}()

	c := make(chan os.Signal, 1)
	// We'll accept graceful shutdowns when quit via SIGINT (Ctrl+C)
	// SIGKILL, SIGQUIT or SIGTERM (Ctrl+/) will not be caught.
	signal.Notify(c, os.Interrupt)

	// Block until we receive our signal.
	<-c

	// Create a deadline to wait for.
	ctx, cancel := context.WithTimeout(context.Background(), (30 * time.Second))
	defer cancel()
	log.Println("shutting down")
	// Doesn't block if no connections, but will otherwise wait
	// until the timeout deadline.
	if err = srv.Shutdown(ctx); err != nil {
		log.Println(err)
	}
	err = db.Close()
	if err != nil {
		log.Println(err)
	}
	os.Exit(0)
}

func createDBDir(p string) error {
	dir := path.Dir(p)
	return os.MkdirAll(dir, os.ModePerm)
}

func IndexHandler(w http.ResponseWriter, r *http.Request) {
	if err := templates.ExecuteTemplate(w, "index.html", nil); err != nil {
		http.Error(w, "error generating json: "+err.Error(), 500)
		return
	}
	// fmt.Fprintf(w, "Home")
}

func CreateHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Create")
}

func SecretHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Get")
}
