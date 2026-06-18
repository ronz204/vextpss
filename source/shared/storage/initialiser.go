package storage

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"gorm.io/gorm"
)

type Initialiser struct {
	dbPath string
}

func NewInitialiser(dbPath string) *Initialiser {
	return &Initialiser{dbPath: dbPath}
}

func (i *Initialiser) DBPath() string { return i.dbPath }

// ================================
// Setup creates the config directory, opens the database, and applies migrations.
// Returns the live connection so the caller can reuse it without a second Open.
// ================================
func (i *Initialiser) Setup(ctx context.Context) (*gorm.DB, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	if err := os.MkdirAll(filepath.Dir(i.dbPath), 0700); err != nil {
		return nil, fmt.Errorf("could not create config directory: %w", err)
	}

	db, err := Open(i.dbPath)
	if err != nil {
		return nil, err
	}

	if err := Migrate(db); err != nil {
		Close(db)
		return nil, err
	}

	return db, nil
}
