package database

import (
	"database/sql"
	"errors"
	"fmt"
	"time"
	"vextpss/source/pkg/models"
)

// ErrNotFound is returned when a requested secret name does not exist.
var ErrNotFound = errors.New("not found")

// Insert persists a SecretRecord to the database.
// The record must already contain encrypted data — this layer never receives plaintext.
func Insert(db *sql.DB, r models.SecretRecord) error {
	const q = `
		INSERT INTO secrets (name, type, salt, nonce, encrypted_payload)
		VALUES (?, ?, ?, ?, ?)`

	_, err := db.Exec(q, r.Name, r.Type, r.Salt, r.Nonce, r.EncryptedPayload)
	if err != nil {
		return fmt.Errorf("insert failed: %w", err)
	}
	return nil
}

// GetByName retrieves a single SecretRecord by its unique name.
// Returns ErrNotFound if no record exists with that name.
func GetByName(db *sql.DB, name string) (*models.SecretRecord, error) {
	const q = `
		SELECT id, name, type, salt, nonce, encrypted_payload, created_at, updated_at
		FROM secrets
		WHERE name = ?`

	row := db.QueryRow(q, name)

	var r models.SecretRecord
	var createdAt, updatedAt string

	err := row.Scan(
		&r.ID, &r.Name, &r.Type,
		&r.Salt, &r.Nonce, &r.EncryptedPayload,
		&createdAt, &updatedAt,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("query failed: %w", err)
	}

	r.CreatedAt, _ = time.Parse("2006-01-02 15:04:05", createdAt)
	r.UpdatedAt, _ = time.Parse("2006-01-02 15:04:05", updatedAt)

	return &r, nil
}

// ListAll retrieves all secrets ordered by name. Only metadata is returned —
// no decryption is performed and no master password is required.
func ListAll(db *sql.DB) ([]models.SecretRecord, error) {
	const q = `SELECT id, name, type, created_at FROM secrets ORDER BY name ASC`

	rows, err := db.Query(q)
	if err != nil {
		return nil, fmt.Errorf("list query failed: %w", err)
	}
	defer rows.Close()

	var records []models.SecretRecord
	for rows.Next() {
		var r models.SecretRecord
		var createdAt string
		if err := rows.Scan(&r.ID, &r.Name, &r.Type, &createdAt); err != nil {
			return nil, fmt.Errorf("scan failed: %w", err)
		}
		r.CreatedAt, _ = time.Parse("2006-01-02 15:04:05", createdAt)
		records = append(records, r)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows error: %w", err)
	}

	return records, nil
}

// DeleteByName removes a secret by name. Returns ErrNotFound if no record exists.
func DeleteByName(db *sql.DB, name string) error {
	const q = `DELETE FROM secrets WHERE name = ?`

	result, err := db.Exec(q, name)
	if err != nil {
		return fmt.Errorf("delete failed: %w", err)
	}

	affected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("could not verify deletion: %w", err)
	}
	if affected == 0 {
		return ErrNotFound
	}

	return nil
}
