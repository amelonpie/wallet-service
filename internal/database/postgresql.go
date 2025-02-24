package database

import (
	"database/sql"
	"fmt"

	_ "github.com/lib/pq" // Postgres driver
)

// Connect connects to the Postgres database and returns a *sql.DB instance.
func (cfg *Config) ConnectPostgre() (*sql.DB, error) {
	db, err := sql.Open("postgres", cfg.PostgreAddr)
	if err != nil {
		return nil, fmt.Errorf("failed to open Postgres DB: %w", err)
	}

	// Test the connection
	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping Postgres DB: %w", err)
	}

	cfg.logger.Info("Connected to PostgreSQL successfully")

	return db, nil
}
