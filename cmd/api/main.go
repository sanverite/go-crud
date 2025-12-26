package main

import (
	"context"
	"database/sql"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	_ "github.com/lib/pq"
)

// A string with application version number. Later this
// will be generated automatically at the build time.
const version = "1.0.0"

// A config struct to hold all the configuration settings
// for this application.
type config struct {
	port int
	env  string
	db   struct {
		dsn          string
		maxOpenConns int
		maxIdleConns int
		maxIdleTime  string
	}
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
	flag.StringVar(&cfg.db.dsn, "db-dsn", os.Getenv("GREENLIGHT_DB_DSN"), "PostgreSQL DSN")
	flag.IntVar(&cfg.db.maxOpenConns, "db-max-open-conns", 25, "PostgreSQL max open connections")
	flag.IntVar(&cfg.db.maxIdleConns, "db-max-idle-conns", 25, "PostgreSQL max idle connections")
	flag.StringVar(&cfg.db.maxIdleTime, "db-max-idle-time", "15m", "PostgreSQL max connection idle time")
	flag.Parse()

	// Initialize a new logger which writes messages to the
	// standard out stream, prefixed with the current timestamp.
	logger := log.New(os.Stdout, "", log.Ldate|log.Ltime)

	// Call the openDB() function to establish DB conn, if it throws
	// an error log and exit the application immediately.
	db, err := openDB(cfg)
	if err != nil {
		logger.Fatal(err)
	}

	// Defer a call to db.Close() so that the connection pool is
	// closed before the main() function exits.
	defer db.Close()

	logger.Printf("database connection pool established")

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
	err = srv.ListenAndServe()
	logger.Fatal(err)
}

// The openDB() function returns a sql.DB connection pool.
func openDB(cfg config) (*sql.DB, error) {
	// Use sql.Open() to create an empty connection pool, using
	// the DSN from the config struct.
	db, err := sql.Open("postgres", cfg.db.dsn)
	if err != nil {
		return nil, err
	}

	// Set the maximum number of open (in-use + idle) connections
	// in the pool.
	db.SetMaxOpenConns(cfg.db.maxOpenConns)

	// Set the maximum number of idle connections in the pool.
	db.SetMaxIdleConns(cfg.db.maxIdleConns)

	// Use the time.ParseDuration() function to convert the idle
	// timeout duration string to a time.Duration type.
	duration, err := time.ParseDuration(cfg.db.maxIdleTime)
	if err != nil {
		return nil, err
	}

	// Set the maximum idle timeout.
	db.SetConnMaxIdleTime(duration)

	// Create a context with a 5-second timeout deadline.
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Use PingContext() to establish a new connection to the database,
	// passing in the context we created above as a parameter. If the
	// conn couldn't be established successfully within the 5 second
	// deadline, then this will return an error.
	err = db.PingContext(ctx)
	if err != nil {
		return nil, err
	}

	return db, nil
}
