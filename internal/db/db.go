package db

import (
	"database/sql"
	"fmt"
	"log"
	"os"

	_ "github.com/lib/pq"
	"github.com/pressly/goose/v3"
)

var DB *sql.DB

func Connect() error {
	var err error

	dbHost := os.Getenv("DB_HOST")
	dbPort := os.Getenv("DB_PORT")
	dbUser := os.Getenv("DB_USER")
	dbPassword := os.Getenv("DB_PASSWORD")
	dbName := os.Getenv("DB_NAME")

	dsn := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		dbHost, dbPort, dbUser, dbPassword, dbName)

	DB, err = sql.Open("postgres", dsn)
	if err != nil {
		return fmt.Errorf("cannot open database: %w", err)
	}

	// Test connection
	err = DB.Ping()
	if err != nil {
		return fmt.Errorf("cannot ping database: %w", err)
	}

	log.Println("✅ Connected to Database!")

	// Run migrations
	err = runMigrations()
	if err != nil {
		return fmt.Errorf("cannot run migrations: %w", err)
	}

	return nil
}

func Close() {
	if DB != nil {
		DB.Close()
	}
}

func runMigrations() error {
	migrationsDir := "./internal/db/migrations"

	if err := goose.SetDialect("postgres"); err != nil {
		return err
	}

	if err := goose.Up(DB, migrationsDir); err != nil {
		return err
	}

	log.Println("✅ Database migrations applied successfully!")
	return nil
}
