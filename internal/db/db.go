package db

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

var DB *sql.DB

func Init(path string) error {
	// Ensure database directory exists
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create database directory: %v", err)
	}

	var err error
	DB, err = sql.Open("sqlite3", path+"?_foreign_keys=on")
	if err != nil {
		return fmt.Errorf("failed to open database: %v", err)
	}

	// Test the connection
	if err = DB.Ping(); err != nil {
		return fmt.Errorf("failed to ping database: %v", err)
	}

	// Set connection pool settings
	DB.SetConnMaxLifetime(time.Minute * 3)
	DB.SetMaxOpenConns(10)
	DB.SetMaxIdleConns(5)

	// Enable foreign keys and WAL mode for better performance
	if _, err := DB.Exec("PRAGMA foreign_keys = ON"); err != nil {
		return fmt.Errorf("failed to enable foreign keys: %v", err)
	}

	if _, err := DB.Exec("PRAGMA journal_mode = WAL"); err != nil {
		return fmt.Errorf("failed to set WAL mode: %v", err)
	}

	// Create tables in the correct order (users first, then books)
	if err := createUsersTable(); err != nil {
		return fmt.Errorf("failed to create users table: %v", err)
	}

	if err := createBooksTable(); err != nil {
		return fmt.Errorf("failed to create books table: %v", err)
	}

	if err := createISBNCacheTable(); err != nil {
		return fmt.Errorf("failed to create ISBN cache table: %v", err)
	}

	if err := createLendingTable(); err != nil {
		return fmt.Errorf("failed to create lending table: %v", err)
	}

	if err := createReadingHistoryTable(); err != nil {
		return fmt.Errorf("failed to create reading history table: %v", err)
	}

	if err := createSettingsTable(); err != nil {
		return fmt.Errorf("failed to create settings table: %v", err)
	}

	if err := createUserSettingsTable(); err != nil {
		return fmt.Errorf("failed to create user settings table: %v", err)
	}

	log.Println("Database initialized successfully")
	return nil
}

func createUsersTable() error {
	userSchema := `
	CREATE TABLE IF NOT EXISTS users (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		username TEXT UNIQUE NOT NULL,
		email TEXT UNIQUE NOT NULL,
		password_hash TEXT NOT NULL,
		role TEXT DEFAULT 'user' CHECK (role IN ('user', 'admin')),
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);`

	if _, err := DB.Exec(userSchema); err != nil {
		return err
	}

	// Create indexes for better performance
	indexes := []string{
		"CREATE INDEX IF NOT EXISTS idx_users_username ON users(username);",
		"CREATE INDEX IF NOT EXISTS idx_users_email ON users(email);",
	}

	for _, index := range indexes {
		if _, err := DB.Exec(index); err != nil {
			return fmt.Errorf("failed to create index: %v", err)
		}
	}

	return nil
}

func createBooksTable() error {
	bookSchema := `
	CREATE TABLE IF NOT EXISTS books (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		user_id INTEGER NOT NULL,
		title TEXT NOT NULL,
		author TEXT,
		isbn TEXT,
		genre TEXT,
		read INTEGER DEFAULT 0 CHECK (read IN (0, 1)),
		cover_url TEXT,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
	);`

	if _, err := DB.Exec(bookSchema); err != nil {
		return err
	}

	// Create indexes for better performance
	indexes := []string{
		"CREATE INDEX IF NOT EXISTS idx_books_user_id ON books(user_id);",
		"CREATE INDEX IF NOT EXISTS idx_books_isbn ON books(isbn);",
		"CREATE INDEX IF NOT EXISTS idx_books_title ON books(title);",
		"CREATE INDEX IF NOT EXISTS idx_books_author ON books(author);",
	}

	for _, index := range indexes {
		if _, err := DB.Exec(index); err != nil {
			return fmt.Errorf("failed to create index: %v", err)
		}
	}

	// Create unique constraint on user_id and isbn
	// Check if the unique index already exists before creating it
	var count int
	err := DB.QueryRow(`
		SELECT COUNT(*) FROM sqlite_master 
		WHERE type='index' AND name='idx_books_user_isbn_unique'
	`).Scan(&count)

	if err == nil && count == 0 {
		// Before creating the unique index, check if there are any duplicates
		var dupCount int
		err = DB.QueryRow(`
			SELECT COUNT(*) FROM (
				SELECT user_id, isbn 
				FROM books 
				WHERE isbn IS NOT NULL AND isbn != ''
				GROUP BY user_id, isbn 
				HAVING COUNT(*) > 1
			)
		`).Scan(&dupCount)

		if err == nil && dupCount > 0 {
			log.Printf("Warning: Found %d duplicate ISBN entries for users. These will be preserved, but new duplicates will be prevented.", dupCount)
		}

		// Try to create the unique index, but don't fail if it can't be created due to existing duplicates
		_, err = DB.Exec("CREATE UNIQUE INDEX IF NOT EXISTS idx_books_user_isbn_unique ON books(user_id, isbn) WHERE isbn IS NOT NULL AND isbn != '';")
		if err != nil {
			log.Printf("Warning: Could not create unique index on user_id and isbn (likely due to existing duplicates): %v", err)
			log.Println("Duplicate ISBN prevention will be handled in application code instead")
		}
	}

	return nil
}

func createISBNCacheTable() error {
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
		return err
	}

	// Create index for cache expiration cleanup
	if _, err := DB.Exec("CREATE INDEX IF NOT EXISTS idx_isbn_cache_cached_at ON isbn_cache(cached_at);"); err != nil {
		return fmt.Errorf("failed to create cache index: %v", err)
	}

	return nil
}

func createLendingTable() error {
	lendingSchema := `
	CREATE TABLE IF NOT EXISTS lending (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		book_id INTEGER NOT NULL,
		user_id INTEGER NOT NULL,
		lent_to TEXT NOT NULL,
		lent_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		due_date DATETIME,
		returned_at DATETIME,
		last_reminder_sent DATETIME,
		FOREIGN KEY (book_id) REFERENCES books(id) ON DELETE CASCADE,
		FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
	);`

	if _, err := DB.Exec(lendingSchema); err != nil {
		return err
	}

	// Create indexes for better performance
	indexes := []string{
		"CREATE INDEX IF NOT EXISTS idx_lending_book_id ON lending(book_id);",
		"CREATE INDEX IF NOT EXISTS idx_lending_user_id ON lending(user_id);",
		"CREATE INDEX IF NOT EXISTS idx_lending_returned_at ON lending(returned_at);",
		"CREATE INDEX IF NOT EXISTS idx_lending_due_date ON lending(due_date);",
		"CREATE INDEX IF NOT EXISTS idx_lending_last_reminder_sent ON lending(last_reminder_sent);",
	}

	for _, index := range indexes {
		if _, err := DB.Exec(index); err != nil {
			return fmt.Errorf("failed to create index: %v", err)
		}
	}

	return nil
}

func createReadingHistoryTable() error {
	readingHistorySchema := `
	CREATE TABLE IF NOT EXISTS reading_history (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		book_id INTEGER NOT NULL,
		user_id INTEGER NOT NULL,
		started_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		completed_at DATETIME,
		FOREIGN KEY (book_id) REFERENCES books(id) ON DELETE CASCADE,
		FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
	);`

	if _, err := DB.Exec(readingHistorySchema); err != nil {
		return err
	}

	// Create indexes for better performance
	indexes := []string{
		"CREATE INDEX IF NOT EXISTS idx_reading_history_book_id ON reading_history(book_id);",
		"CREATE INDEX IF NOT EXISTS idx_reading_history_user_id ON reading_history(user_id);",
		"CREATE INDEX IF NOT EXISTS idx_reading_history_completed_at ON reading_history(completed_at);",
		"CREATE INDEX IF NOT EXISTS idx_reading_history_started_at ON reading_history(started_at);",
	}

	for _, index := range indexes {
		if _, err := DB.Exec(index); err != nil {
			return fmt.Errorf("failed to create index: %v", err)
		}
	}

	return nil
}

func createSettingsTable() error {
	settingsSchema := `
	CREATE TABLE IF NOT EXISTS settings (
		key TEXT PRIMARY KEY,
		value TEXT NOT NULL,
		updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);`

	if _, err := DB.Exec(settingsSchema); err != nil {
		return err
	}

	// Insert default settings if they don't exist
	_, err := DB.Exec(`
		INSERT OR IGNORE INTO settings (key, value) 
		VALUES ('registration_enabled', 'true')
	`)
	if err != nil {
		return fmt.Errorf("failed to insert default settings: %v", err)
	}

	return nil
}

func createUserSettingsTable() error {
	userSettingsSchema := `
	CREATE TABLE IF NOT EXISTS user_settings (
		user_id INTEGER PRIMARY KEY,
		email_reminders_enabled BOOLEAN DEFAULT 1,
		email_upcoming_reminders BOOLEAN DEFAULT 1,
		email_overdue_reminders BOOLEAN DEFAULT 1,
		default_lending_days INTEGER DEFAULT 14,
		yearly_reading_goal INTEGER DEFAULT 0,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
	);`

	if _, err := DB.Exec(userSettingsSchema); err != nil {
		return err
	}

	return nil
}

// Close closes the database connection
func Close() error {
	if DB != nil {
		return DB.Close()
	}
	return nil
}

// GetDB returns the database instance
func GetDB() *sql.DB {
	return DB
}
