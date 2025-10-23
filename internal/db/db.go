package db

import (
	"database/sql"
	"log"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

var DB *sql.DB

func Init(path string) {
	var err error
	DB, err = sql.Open("sqlite3", path)
	if err != nil {
		log.Fatalf("failed to open db: %v", err)
	}

	DB.SetConnMaxLifetime(time.Minute * 3)
	DB.SetMaxOpenConns(1)
	DB.SetMaxIdleConns(1)

	bookSchema := `
	CREATE TABLE IF NOT EXISTS books (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		title TEXT NOT NULL,
		author TEXT,
		isbn TEXT,
		genre TEXT,
		read INTEGER DEFAULT 0,
		lent_to TEXT,
		lent_at DATETIME
	);`
	if _, err := DB.Exec(bookSchema); err != nil {
		log.Fatalf("failed to create schema: %v", err)
	}

	cacheSchema := `
	CREATE TABLE IF NOT EXISTS isbn_cache (
		isbn TEXT PRIMARY KEY,
		title TEXT,
		author TEXT,
		genre TEXT,
		cover_url TEXT,
		cached_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);`
	if _, err := DB.Exec(cacheSchema); err != nil {
		log.Fatalf("failed to create schema: %v", err)
	}
}
