// Package db provides a shared PostgreSQL connection pool for the Health Data Platform.
// Uses pgx stdlib driver with database/sql interface for compatibility and ergonomic pooling.
package db

import (
	"database/sql"
	"fmt"

	_ "github.com/jackc/pgx/v5/stdlib" // register pgx driver
)

// NewPool opens a *sql.DB connection pool using the pgx driver.
// Caller is responsible for defering pool.Close().
func NewPool(databaseURL string) (*sql.DB, error) {
	db, err := sql.Open("pgx", databaseURL)
	if err != nil {
		return nil, fmt.Errorf("db: open: %w", err)
	}

	// Tuned for a low-traffic health platform (not a high-throughput service)
	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(5)

	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("db: ping: %w", err)
	}

	return db, nil
}
