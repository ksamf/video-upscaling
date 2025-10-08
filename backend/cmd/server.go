package main

import (
	"fmt"
	"log"
	"net/http"
)

func (app *application) serve() error {
	server := &http.Server{
		Addr:    fmt.Sprintf("%s:%d", app.host, app.port),
		Handler: app.routes(),
	}
	log.Printf("Starting server on port %d", app.port)

	return server.ListenAndServe()
}
