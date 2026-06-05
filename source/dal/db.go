package dal

import (
	"fmt"

	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// Open opens a GORM connection to the SQLite database at path.
// Logging is silenced so SQL never reaches the user's terminal.
func Open(path string) (*gorm.DB, error) {
	db, err := gorm.Open(sqlite.Open(path), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		return nil, fmt.Errorf("could not open database: %w", err)
	}
	return db, nil
}

// Close releases the underlying sql.DB connection pool.
func Close(db *gorm.DB) error {
	sqlDB, err := db.DB()
	if err != nil {
		return fmt.Errorf("could not access underlying db: %w", err)
	}
	return sqlDB.Close()
}

// Migrate runs AutoMigrate to create or update the secrets table.
// Safe to call multiple times — GORM only adds missing columns/indexes.
func Migrate(db *gorm.DB) error {
	if err := db.AutoMigrate(&SecretRecord{}); err != nil {
		return fmt.Errorf("migration failed: %w", err)
	}
	return nil
}
