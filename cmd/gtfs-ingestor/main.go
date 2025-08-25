package main

import (
	"context"
	"database/sql"
	"log"
	"os"
	"time"

	"arrivo-transit-api/internal/database"
	"arrivo-transit-api/internal/gtfs"
	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	_ "github.com/jackc/pgx/v5/stdlib"
)

func main() {
	log.Println("Starting GTFS ingestor...")

	ctx := context.Background()

	// Get database connection string from environment variable
	dsn := os.Getenv("POSTGRES_DSN")
	if dsn == "" {
		log.Fatal("POSTGRES_DSN environment variable not set")
	}

	// Create a new database connection pool.
	pool, err := database.NewDB(ctx, dsn)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer pool.Close()

	// Run database migrations using a standard library connection
	sqlDB, err := sql.Open("pgx", dsn)
	if err != nil {
		log.Fatalf("Failed to open a connection for migrations: %v", err)
	}
	defer sqlDB.Close()

	if err := runMigrations(sqlDB); err != nil {
		log.Fatalf("Failed to run migrations: %v", err)
	}

	gtfsService := gtfs.NewService(pool)

	for {
		if err := gtfsService.IngestGTFSData(); err != nil {
			log.Printf("Error ingesting GTFS data: %v", err)
		}

		log.Println("GTFS ingestor sleeping for 1 hour...")
		time.Sleep(1 * time.Hour)
	}
}

func runMigrations(db *sql.DB) error {
	driver, err := postgres.WithInstance(db, &postgres.Config{})
	if err != nil {
		return err
	}

	m, err := migrate.NewWithDatabaseInstance(
		"file:///app/internal/database/migrations",
		dbName(os.Getenv("POSTGRES_DSN")),
		driver,
	)
	if err != nil {
		return err
	}

	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		return err
	}

	log.Println("Database migrations applied successfully")
	return nil
}

// dbName extracts the database name from the DSN string.
func dbName(dsn string) string {
	// This is a simplified parser. A more robust solution would handle more complex DSNs.
	// Example DSN: "postgres://user:password@host:port/dbname?sslmode=disable"
	// We want to extract "dbname".
	lastSlash := 0
	for i, r := range dsn {
		if r == '/' {
			lastSlash = i
		}
	}
	qMark := len(dsn)
	for i, r := range dsn {
		if r == '?' {
			qMark = i
			break
		}
	}

	if lastSlash+1 > qMark {
		return ""
	}

	return dsn[lastSlash+1 : qMark]
}