package db

import (
	"database/sql"
	"embed"
	"os"
	"path/filepath"
	"strings"

	"github.com/pressly/goose/v3"

	_ "github.com/mattn/go-sqlite3"
	_ "github.com/tursodatabase/libsql-client-go/libsql"
)

//go:embed migrations/*.sql
var embedMigrations embed.FS

func Open(dbPath string) (*sql.DB, error) {
	var db *sql.DB
	var err error

	if strings.HasPrefix(dbPath, "libsql://") {
		// Turso connection
		db, err = sql.Open("libsql", dbPath)
	} else {
		// Local SQLite file
		if dir := filepath.Dir(dbPath); dir != "." && dir != "" {
			if err := os.MkdirAll(dir, 0755); err != nil {
				return nil, err
			}
		}
		dsn := dbPath + "?_foreign_keys=on"
		db, err = sql.Open("sqlite3", dsn)
	}

	if err != nil {
		return nil, err
	}

	// Test connection
	if err := db.Ping(); err != nil {
		return nil, err
	}

	return db, nil
}

func Migrate(db *sql.DB) error {
	goose.SetBaseFS(embedMigrations)

	if err := goose.SetDialect("sqlite3"); err != nil {
		return err
	}

	return goose.Up(db, "migrations")
}
