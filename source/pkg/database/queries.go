package database

import (
	"errors"
	"fmt"

	"vextpss/source/pkg/models"

	"gorm.io/gorm"
)

// ErrNotFound is returned when a requested secret name does not exist.
var ErrNotFound = errors.New("not found")

// Insert persists a SecretRecord to the database.
// The record must already contain encrypted data — this layer never receives plaintext.
func Insert(db *gorm.DB, r models.SecretRecord) error {
	result := db.Create(&r)
	if result.Error != nil {
		return fmt.Errorf("insert failed: %w", result.Error)
	}
	return nil
}

// GetByName retrieves a single SecretRecord by its unique name.
// Returns ErrNotFound if no record exists with that name.
func GetByName(db *gorm.DB, name string) (*models.SecretRecord, error) {
	var r models.SecretRecord
	result := db.Where("name = ?", name).First(&r)
	if errors.Is(result.Error, gorm.ErrRecordNotFound) {
		return nil, ErrNotFound
	}
	if result.Error != nil {
		return nil, fmt.Errorf("query failed: %w", result.Error)
	}
	return &r, nil
}

// ListAll retrieves all secrets ordered by name. Only metadata is returned —
// no decryption is performed and no master password is required.
func ListAll(db *gorm.DB) ([]models.SecretRecord, error) {
	var records []models.SecretRecord
	result := db.Select("id, name, type, created_at").Order("name asc").Find(&records)
	if result.Error != nil {
		return nil, fmt.Errorf("list query failed: %w", result.Error)
	}
	return records, nil
}

// DeleteByName removes a secret by name. Returns ErrNotFound if no record exists.
func DeleteByName(db *gorm.DB, name string) error {
	result := db.Where("name = ?", name).Delete(&models.SecretRecord{})
	if result.Error != nil {
		return fmt.Errorf("delete failed: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return ErrNotFound
	}
	return nil
}
