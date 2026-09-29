package db

import (
	"database/sql"
	"fmt"

	_ "github.com/jackc/pgx/v5/stdlib"
)

// Connect opens a PostgreSQL connection using the pgx driver.
func Connect(dsn string) (*sql.DB, error) {
	database, err := sql.Open("pgx", dsn)
	if err != nil {
		return nil, fmt.Errorf("open database connection: %w", err)
	}

	if err := database.Ping(); err != nil {
		_ = database.Close()
		return nil, fmt.Errorf("ping database: %w", err)
	}

	return database, nil
}
