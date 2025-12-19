package main

import (
	"log"

	"chess-backend/internal/config"
)

type app struct {
	app_name   string
	jwt_secret string
	port       int
	// database_models

}

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Config load error: %v", err)
	}

	app := &app{
		app_name:   "Go-Chess",
		jwt_secret: "sonme-secret-key",
		port:       cfg.Server.Port,
	}

	if err := app.serve(); err != nil {
		log.Fatalf("Serve error %v", err)
	}
}
