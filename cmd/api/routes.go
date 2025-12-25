package main

import (
	"net/http"

	"github.com/julienschmidt/httprouter"
)

func (app *application) routes() *httprouter.Router {
	// Initialize a new httprouter router instance.
	router := httprouter.New()

	// Convert the notFoundResponse() helper to a http.Handler using
	// the http.HandlerFunc() adapted, and the set it as the custom
	// error handler for 404 Not Found responses.
	router.NotFound = http.HandlerFunc(app.notFoundResponse)

	router.MethodNotAllowed = http.HandlerFunc(app.methodNotAllowedResponse)

	// Register relevant methods, URL paterns and handler functions
	// for the endpoints using the HandleFunc() method.
	router.HandlerFunc(http.MethodGet, "/v1/health", app.healthcheckHandler)
	router.HandlerFunc(http.MethodGet, "/v1/healthcheck", app.healthcheckJSONHandler)
	router.HandlerFunc(http.MethodPost, "/v1/movies", app.createMovieHandler)
	router.HandlerFunc(http.MethodGet, "/v1/movies/:id", app.showMovieHandler)

	// Return the httprouter instance
	return router
}
