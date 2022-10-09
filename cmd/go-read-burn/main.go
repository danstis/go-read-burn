package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"path"
	"time"

	"github.com/boltdb/bolt"
	"github.com/danstis/go-read-burn/internal/version"
	"github.com/gorilla/mux"
)

var (
	dbPath = path.Join("..", "..", "db", "secrets.db")
	db     *bolt.DB
)

// Main entry point for the app.
func main() {
	log.Printf("Version %q", version.Version)

	var err error
	db, err = bolt.Open(dbPath, 0644, nil)
	if err != nil {
		log.Println(err)
	}
	defer db.Close()

	r := mux.NewRouter()
	r.HandleFunc("/", IndexHandler)
	r.HandleFunc("/create", CreateHandler).Methods("POST")
	r.HandleFunc("/get/{key}", SecretHandler)
	http.Handle("/", r)

	srv := &http.Server{
		Handler:      r,
		Addr:         "0.0.0.0:8080",
		WriteTimeout: 15 * time.Second,
		ReadTimeout:  15 * time.Second,
	}

	// Run our server in a goroutine so that it doesn't block.
	go func() {
		log.Printf("Started server listing on %s", "0.0.0.0:8080")
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
	srv.Shutdown(ctx)
	err = db.Close()
	if err != nil {
		log.Println(err)
	}
	os.Exit(0)
}

func IndexHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Home")
}

func CreateHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Create")
}

func SecretHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Get")
}
