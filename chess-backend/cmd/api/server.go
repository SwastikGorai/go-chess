package main

import (
	"fmt"
	"log"
	"net/http"
	"time"
)

func (app *app) serve() error {
	server := http.Server{
		Addr:         fmt.Sprintf("localhost:%d", app.port),
		Handler:      app.routes(),
		IdleTimeout:  time.Minute,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 30 * time.Second,
	}

	log.Printf("Starting server on port %d", app.port)

	return server.ListenAndServe()
}
