package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"os"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"
)

type DB struct {
	*sql.DB
}

// NewDB initializes a PostgreSQL connection pool, auto-creates database if missing, and runs schema migrations
func NewDB(databaseURL string) (*DB, error) {
	db, err := sql.Open("pgx", databaseURL)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(10)
	db.SetConnMaxLifetime(5 * time.Minute)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := db.PingContext(ctx); err != nil {
		// If DB doesn't exist, try creating it via default maintenance postgres db
		if tryCreateDatabase(databaseURL) {
			// Retry connecting
			db, err = sql.Open("pgx", databaseURL)
			if err == nil && db.PingContext(ctx) == nil {
				pgDB := &DB{db}
				_ = pgDB.runAutoMigrations()
				return pgDB, nil
			}
		}
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	pgDB := &DB{db}

	if err := pgDB.runAutoMigrations(); err != nil {
		log.Printf("Warning: Failed running auto migrations: %v", err)
	}

	return pgDB, nil
}

func tryCreateDatabase(databaseURL string) bool {
	// Parse database name from URL e.g. postgres://user:pass@host:port/dbname?sslmode=disable
	// Connect to /postgres instead and execute CREATE DATABASE dbname
	baseURL, dbName := extractDBName(databaseURL)
	if dbName == "" || dbName == "postgres" {
		return false
	}

	maintenanceURL := baseURL + "/postgres?sslmode=disable"
	mDb, err := sql.Open("pgx", maintenanceURL)
	if err != nil {
		return false
	}
	defer mDb.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Safe identifier check
	_, _ = mDb.ExecContext(ctx, fmt.Sprintf(`CREATE DATABASE "%s"`, dbName))
	log.Printf("Created database '%s' automatically.", dbName)
	return true
}

func extractDBName(rawURL string) (string, string) {
	// Simplified parsing for standard postgresql URLs
	var base, name string
	idx := len(rawURL) - 1
	for idx >= 0 && rawURL[idx] != '/' {
		idx--
	}
	if idx >= 0 {
		base = rawURL[:idx]
		rest := rawURL[idx+1:]
		if qIdx := len(rest); qIdx > 0 {
			for i, ch := range rest {
				if ch == '?' {
					name = rest[:i]
					return base, name
				}
			}
			name = rest
		}
	}
	return base, name
}

func (db *DB) runAutoMigrations() error {
	// Try reading migration file first
	migrationFile := "migrations/001_init.sql"
	content, err := os.ReadFile(migrationFile)
	var query string
	if err == nil {
		query = string(content)
	} else {
		// Fallback inline migration
		query = `
		CREATE TABLE IF NOT EXISTS toll_plazas (
			id SERIAL PRIMARY KEY,
			name VARCHAR(255) NOT NULL,
			latitude DOUBLE PRECISION NOT NULL,
			longitude DOUBLE PRECISION NOT NULL,
			geo_state VARCHAR(100),
			created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
			CONSTRAINT uq_toll_plaza UNIQUE (name, latitude, longitude)
		);

		CREATE INDEX IF NOT EXISTS idx_toll_plazas_lat ON toll_plazas(latitude);
		CREATE INDEX IF NOT EXISTS idx_toll_plazas_lon ON toll_plazas(longitude);
		CREATE INDEX IF NOT EXISTS idx_toll_plazas_lat_lon ON toll_plazas(latitude, longitude);
		CREATE INDEX IF NOT EXISTS idx_toll_plazas_state ON toll_plazas(geo_state);

		CREATE TABLE IF NOT EXISTS pincode_cache (
			pincode VARCHAR(10) PRIMARY KEY,
			latitude DOUBLE PRECISION NOT NULL,
			longitude DOUBLE PRECISION NOT NULL,
			display_name TEXT,
			created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
		);
		`
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	_, err = db.ExecContext(ctx, query)
	return err
}
