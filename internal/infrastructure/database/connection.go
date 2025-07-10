/*
Package database provides utilities for initializing and managing
database connections for the application.

This package currently supports PostgreSQL connection initialization
with a sane default connection pool configuration.
*/
package database

import (
	"database/sql"
	"fmt"
	"time"

	_ "github.com/lib/pq" // PostgreSQL driver
	"log-parser/internal/config"
)

// NewPostgresDB initializes and returns a PostgreSQL database connection
// configured with optimal connection pool settings.
//
// The function performs the following steps:
//  1. Opens a new connection to the PostgreSQL database using the DSN
//     provided by the configuration.
//  2. Configures the connection pool with:
//     - Max open connections: 25
//     - Max idle connections: 25
//     - Connection max lifetime: 5 minutes
//  3. Pings the database to ensure the connection is valid.
//
// Parameters:
//   - cfg: pointer to DatabaseConfig which provides the DSN string.
//
// Returns:
//   - *sql.DB: a live database connection ready for queries.
//   - error: if opening or pinging the database fails.
//
// Example usage:
//
//	db, err := database.NewPostgresDB(myConfig)
//	if err != nil {
//	    log.Fatalf("failed to connect to DB: %v", err)
//	}
//	defer db.Close()
func NewPostgresDB(cfg *config.DatabaseConfig) (*sql.DB, error) {
	db, err := sql.Open("postgres", cfg.DSN())
	if err != nil {
		fmt.Printf("failed to open DB with DSN: %v\n", cfg.DSN())
		return nil, err
	}

	// Configure connection pool parameters
	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(25)
	db.SetConnMaxLifetime(5 * time.Minute)

	// Verify the database connection is alive
	if err := db.Ping(); err != nil {
		_ = db.Close() // Ensure DB handle is closed if ping fails
		return nil, fmt.Errorf("unable to ping database: %w", err)
	}

	return db, nil
}
