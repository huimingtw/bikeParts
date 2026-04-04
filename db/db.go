package db

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"

	_ "github.com/mattn/go-sqlite3"
)

type DB struct {
	Db      *sql.DB
	logFile string // path to log file
}

func Init() (*DB, error) {
	var DB_PATH = os.Getenv("DB_PATH")
	if DB_PATH == "" {
		DB_PATH = "./db/data.db"
	}

	if err := os.MkdirAll(filepath.Dir(DB_PATH), 0755); err != nil {
		return nil, err
	}

	db, err := sql.Open("sqlite3", DB_PATH)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %v", err)
	}
	db.SetMaxOpenConns(1) // SQLite does not support concurrent writes, so we limit to 1 connection

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

	logPath := os.Getenv("LOG_PATH")
	if logPath == "" {
		logPath = "."
	}
	logFile := filepath.Join(filepath.Dir(logPath), "app.log")
	d := &DB{Db: db, logFile: logFile}

	count := 0
	db.QueryRow("SELECT COUNT(*) FROM parts").Scan(&count)
	if count == 0 {
		if err := d.seedInitialData(); err != nil {
			return nil, fmt.Errorf("failed to seed initial data: %v", err)
		}
	}

	return d, nil
}

func (d *DB) seedInitialData() error {
	seedPath := os.Getenv("SEED_PATH")
	if seedPath == "" {
		seedPath = "./db/seed.sql"
	}
	seedBytes, err := os.ReadFile(seedPath)
	if err != nil {
		return fmt.Errorf("failed to read seed file: %v", err)
	}
	if _, err := d.Db.Exec(string(seedBytes)); err != nil {
		return fmt.Errorf("failed to execute seed data: %v", err)
	}
	return nil
}

func (d *DB) Close() error {
	return d.Db.Close()
}
