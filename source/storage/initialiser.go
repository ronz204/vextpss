package storage

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
)

// Initialiser manages the one-time database setup needed by InitStorageUC.
type Initialiser struct {
	dbPath string
}

func NewInitialiser(dbPath string) *Initialiser {
	return &Initialiser{dbPath: dbPath}
}

func (i *Initialiser) DBPath() string { return i.dbPath }

// Init creates the app directory, opens/creates the database, and applies migrations.
func (i *Initialiser) Init(ctx context.Context) error {
	if err := os.MkdirAll(filepath.Dir(i.dbPath), 0700); err != nil {
		return fmt.Errorf("could not create config directory: %w", err)
	}
	db, err := Open(i.dbPath)
	if err != nil {
		return err
	}
	defer Close(db)
	return Migrate(db)
}
