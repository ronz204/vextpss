package database

import (
	"database/sql"
	"fmt"

	_ "modernc.org/sqlite"
)

const schema = `
CREATE TABLE IF NOT EXISTS secrets (
    id                INTEGER  PRIMARY KEY AUTOINCREMENT,
    name              TEXT     UNIQUE NOT NULL,
    type              TEXT     NOT NULL,
    salt              BLOB     NOT NULL,
    nonce             BLOB     NOT NULL,
    encrypted_payload BLOB     NOT NULL,
    created_at        DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at        DATETIME DEFAULT CURRENT_TIMESTAMP
);`

// Open opens a connection to the SQLite database at the given path.
// The caller is responsible for closing the returned *sql.DB.
func Open(path string) (*sql.DB, error) {
	db, err := sql.Open("sqlite", path)
	if err != nil {
		return nil, fmt.Errorf("could not open database: %w", err)
	}
	if err := db.Ping(); err != nil {
		db.Close()
		return nil, fmt.Errorf("could not connect to database: %w", err)
	}
	return db, nil
}

// Migrate applies the schema to the database. It is safe to call multiple times
// because it uses CREATE TABLE IF NOT EXISTS.
func Migrate(db *sql.DB) error {
	if _, err := db.Exec(schema); err != nil {
		return fmt.Errorf("migration failed: %w", err)
	}
	return nil
}
