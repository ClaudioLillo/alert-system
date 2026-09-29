// @title Alert System API
// @version 1.0
// @description A thin API layer that delegates business rules to PostgreSQL functions and triggers.
// @host localhost:8080
// @BasePath /
package main

import (
	"context"
	"log"
	"net/http"

	"alert-system/internal/api"
	"alert-system/internal/config"
	"alert-system/internal/db"
)

func main() {
	cfg := config.Load()

	database, err := db.Connect(cfg.DSN())
	if err != nil {
		log.Fatalf("failed to connect to PostgreSQL: %v", err)
	}
	defer func() {
		if cerr := database.Close(); cerr != nil {
			log.Printf("close database connection: %v", cerr)
		}
	}()

	if err := db.RunMigrations(context.Background(), database, cfg.MigrationDir); err != nil {
		log.Fatalf("failed to apply database migrations: %v", err)
	}

	handler := api.NewRouter(database)
	log.Printf("starting HTTP server on %s", cfg.HTTPAddr)
	if err := http.ListenAndServe(cfg.HTTPAddr, handler); err != nil {
		log.Fatalf("http server failed: %v", err)
	}
}
