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
	"github.com/danstis/go-read-burn/internal/storage"
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

	config, err := loadConfig()
	if err != nil {
		log.Fatalf("failed to load config: %v", err)
	}

	db, err = openDB(config.DBPath)
	if err != nil {
		log.Fatalf("failed to open DB: %v", err)
	}
	defer db.Close()

	if err := storage.InitBucket(db); err != nil {
		log.Fatalf("failed to init secrets bucket: %v", err)
	}

	r := mux.NewRouter()
	setupRoutes(r)

	templates, err = parseTemplates()
	if err != nil {
		log.Fatalf("failed to parse templates: %v", err)
	}

	srv := createServer(config.ListenHost, config.ListenPort, r)

	startServer(srv)

	shutdownServer(srv, db)
}

func loadConfig() (Config, error) {
	var config Config
	err := envconfig.Process("GRB", &config)
	if err != nil {
		return config, err
	}
	return config, nil
}

func openDB(dbPath string) (*bolt.DB, error) {
	err := createDBDir(dbPath)
	if err != nil {
		return nil, fmt.Errorf("failed to create database directory: %w", err)
	}
	db, err := bolt.Open(dbPath, 0644, nil)
	if err != nil {
		return nil, err
	}
	return db, nil
}

func setupRoutes(r *mux.Router) {
	r.HandleFunc("/", IndexHandler)
	r.HandleFunc("/create", CreateHandler).Methods("POST")
	r.HandleFunc("/get/{key}", SecretHandler)
	s := http.StripPrefix("/static/", http.FileServer(http.FS(static)))
	r.PathPrefix("/static/").Handler(s)
	http.Handle("/", r)
}

func parseTemplates() (*template.Template, error) {
	templates, err := template.ParseFS(views, "views/*.html")
	if err != nil {
		return nil, err
	}
	return templates, nil
}

func createServer(listenHost, listenPort string, r *mux.Router) *http.Server {
	srv := &http.Server{
		Handler:      r,
		Addr:         fmt.Sprintf("%s:%s", listenHost, listenPort),
		WriteTimeout: 15 * time.Second,
		ReadTimeout:  15 * time.Second,
	}
	return srv
}

func startServer(srv *http.Server) {
	go func() {
		log.Printf("Started server listing on %s", srv.Addr)
		if err := srv.ListenAndServe(); err != nil {
			log.Println(err)
		}
	}()
}

func shutdownServer(srv *http.Server, db *bolt.DB) {
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	<-c

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	log.Println("shutting down")

	if err := srv.Shutdown(ctx); err != nil {
		log.Println(err)
	}

	err := db.Close()
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
}

func CreateHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Create")
}

func SecretHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Get")
}
