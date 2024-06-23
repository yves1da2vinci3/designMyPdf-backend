package auth

import (
	"database/sql"
	"log"
	"time"

	"github.com/gofiber/storage/sqlite3"
	_ "github.com/mattn/go-sqlite3"
)

// InitSessionDB initializes the SQLite database for session storage.
func InitSessionDB() {
	db, err := sql.Open("sqlite3", "./config/session.db")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	// Ensure the table structure for sessions
	query := `CREATE TABLE IF NOT EXISTS sessions (
		k  VARCHAR(64) PRIMARY KEY NOT NULL DEFAULT '',
		v  BLOB NOT NULL,
		e  BIGINT NOT NULL DEFAULT '0',
		u  TEXT);`
	_, err = db.Exec(query)
	if err != nil {
		log.Fatal(err)
	}
}

// GetSessionStore initializes and returns the session storage.
func GetSessionStore() *sqlite3.Storage {
	storage := sqlite3.New(sqlite3.Config{
		Database:        "./config/session.db",
		Table:           "sessions",
		Reset:           false,
		GCInterval:      10 * time.Second,
		MaxOpenConns:    100,
		MaxIdleConns:    100,
		ConnMaxLifetime: 1 * time.Hour,
	})
	return storage
}
