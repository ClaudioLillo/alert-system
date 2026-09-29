package main

import (
	"context"
	"log"

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
		log.Fatalf("failed to run database migrations: %v", err)
	}

	log.Println("database migrations applied successfully")
}
