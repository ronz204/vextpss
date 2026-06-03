package database

import (
	"fmt"

	"vextpss/source/pkg/models"

	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// Open opens a GORM connection to the SQLite database at the given path.
// Logger is set to Silent so SQL never leaks to the CLI user's stdout.
// The caller is responsible for closing the underlying *sql.DB via db.DB().
func Open(path string) (*gorm.DB, error) {
	db, err := gorm.Open(sqlite.Open(path), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		return nil, fmt.Errorf("could not open database: %w", err)
	}
	return db, nil
}

// Migrate runs AutoMigrate to create or update the secrets table schema.
// It is safe to call multiple times — GORM only adds missing columns/indexes.
func Migrate(db *gorm.DB) error {
	if err := db.AutoMigrate(&models.SecretRecord{}); err != nil {
		return fmt.Errorf("migration failed: %w", err)
	}
	return nil
}
