package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"
)

// A string with application version number. Later this
// will be generated automatically at the build time.
const version = "1.0.0"

// A config struct to hold all the configuration settings
// for this application.
type config struct {
	port int
	env  string
}

// An application struct to hold the dependencies for HTTP
// handlers, helpers, and middlewares.
type application struct {
	config config
	logger *log.Logger
}

func main() {
	// An instance of the config struct
	var cfg config

	// Read the value of the port and env command-line flags
	// into the config struct.
	flag.IntVar(&cfg.port, "port", 4000, "API server port")
	flag.StringVar(&cfg.env, "env", "development", "Environment (development|staging|production)")
	flag.Parse()

	// Initialize a new logger which writes messages to the
	// standard out stream, prefixed with the current timestamp.
	logger := log.New(os.Stdout, "", log.Ldate|log.Ltime)

	// Declare an instance of the application struct, containing
	// the config struct and the logger
	app := &application{
		config: cfg,
		logger: logger,
	}

	// HTTP server with timeout settings, listens on the port provided
	// in the config
	srv := &http.Server{
		Addr:         fmt.Sprintf(":%d", cfg.port),
		Handler:      app.routes(),
		IdleTimeout:  time.Minute,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 30 * time.Second,
	}

	// Start the HTTP server
	logger.Printf("starting %s server on %s\n", cfg.env, srv.Addr)
	err := srv.ListenAndServe()
	logger.Fatal(err)
}
