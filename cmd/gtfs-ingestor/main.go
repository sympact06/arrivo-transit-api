package main

import (
	"arrivo-transit-api/internal/config"
	"context"
	"fmt"
	"log"
	"net/url"
	"strings"

	"github.com/jackc/pgx/v5/pgxpool"
)

func main() {
	cfg := config.Load()

	// Connect to PostgreSQL
	db, err := pgxpool.New(context.Background(), cfg.PostgresDSN)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	// Test the connection
	if err := db.Ping(context.Background()); err != nil {
		log.Fatalf("Failed to ping database: %v", err)
	}

	dbName := dbName(cfg.PostgresDSN)
	log.Printf("Connected to PostgreSQL database: %s", dbName)

	// TODO: Implement GTFS data ingestion
	// This would typically involve:
	// 1. Downloading GTFS files from transit agencies
	// 2. Parsing the GTFS data (stops.txt, routes.txt, etc.)
	// 3. Inserting/updating the data in PostgreSQL
	// 4. Creating appropriate indexes for performance

	log.Println("GTFS ingestor started successfully")
}

// dbName extracts the database name from a PostgreSQL DSN
func dbName(dsn string) string {
	u, err := url.Parse(dsn)
	if err != nil {
		return "unknown"
	}
	return strings.TrimPrefix(u.Path, "/")
}