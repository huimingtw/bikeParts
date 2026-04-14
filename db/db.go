package db

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"

	_ "github.com/mattn/go-sqlite3"
)

type Config struct {
	DBPath string
	Schema []byte // provided by main package
	Seed   []byte // optional seed data for development or tests
}

func Init(cfg Config) (*sql.DB, error) {
	dbPath := cfg.DBPath
	if dbPath == "" {
		return nil, fmt.Errorf("db path is required")
	}

	if dbPath != ":memory:" {
		if err := os.MkdirAll(filepath.Dir(dbPath), 0o755); err != nil {
			return nil, err
		}
	}

	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %v", err)
	}
	db.SetMaxOpenConns(1) // SQLite does not support concurrent writes

	if len(cfg.Schema) == 0 {
		return nil, fmt.Errorf("schema is required")
	}
	if _, err := db.Exec(string(cfg.Schema)); err != nil {
		return nil, fmt.Errorf("failed to execute schema: %v", err)
	}

	if len(cfg.Seed) > 0 {
		count := 0
		if err := db.QueryRow("SELECT COUNT(*) FROM parts").Scan(&count); err != nil {
			return nil, fmt.Errorf("failed to check parts count: %v", err)
		}
		if count == 0 {
			if _, err := db.Exec(string(cfg.Seed)); err != nil {
				return nil, fmt.Errorf("failed to execute seed data: %v", err)
			}
		}
	}

	return db, nil
}
