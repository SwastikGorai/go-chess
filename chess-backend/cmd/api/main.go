package main

import (
	"context"
	"log"

	"chess-backend/internal/config"
	"chess-backend/internal/store"
)

type app struct {
	app_name   string
	jwt_secret string
	port       int
	store      store.GameStore
	// database_models

}

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Config load error: %v", err)
	}

	ctx := context.Background()
	dbStore, err := store.NewPostgresStore(ctx, cfg.Database.DSN())
	if err != nil {
		log.Fatalf("Database connection error: %v", err)
	}
	defer dbStore.Close()

	app := &app{
		app_name:   "Go-Chess",
		jwt_secret: "sonme-secret-key",
		port:       cfg.Server.Port,
		store:      dbStore,
	}

	if err := app.serve(); err != nil {
		log.Fatalf("Serve error %v", err)
	}
}
