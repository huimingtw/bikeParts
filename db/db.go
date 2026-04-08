package db

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"

	_ "github.com/mattn/go-sqlite3"
)

func Init() (*sql.DB, error) {
	dbPath := os.Getenv("DB_PATH")
	if dbPath == "" {
		dbPath = "./db/data.db"
	}

	if err := os.MkdirAll(filepath.Dir(dbPath), 0755); err != nil {
		return nil, err
	}

	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %v", err)
	}
	db.SetMaxOpenConns(1) // SQLite does not support concurrent writes

	schemaPath := os.Getenv("SCHEMA_PATH")
	if schemaPath == "" {
		schemaPath = "./db/schema.sql"
	}
	schemaBytes, err := os.ReadFile(schemaPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read schema file: %v", err)
	}
	if _, err := db.Exec(string(schemaBytes)); err != nil {
		return nil, fmt.Errorf("failed to execute schema: %v", err)
	}

	count := 0
	db.QueryRow("SELECT COUNT(*) FROM parts").Scan(&count)
	if count == 0 {
		if err := seedInitialData(db); err != nil {
			return nil, fmt.Errorf("failed to seed initial data: %v", err)
		}
	}

	return db, nil
}

func seedInitialData(db *sql.DB) error {
	seedPath := os.Getenv("SEED_PATH")
	if seedPath == "" {
		seedPath = "./db/seed.sql"
	}
	seedBytes, err := os.ReadFile(seedPath)
	if err != nil {
		return fmt.Errorf("failed to read seed file: %v", err)
	}
	if _, err := db.Exec(string(seedBytes)); err != nil {
		return fmt.Errorf("failed to execute seed data: %v", err)
	}
	return nil
}
