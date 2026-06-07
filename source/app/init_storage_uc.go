package app

import (
	"context"
	"fmt"
	"os"
)

// StorageInitialiser knows how to prepare the persistence layer.
// Implemented by dal so the use case stays framework-agnostic.
type StorageInitialiser interface {
	// Init creates required directories, the database file, and applies migrations.
	Init(ctx context.Context) error
	// DBPath returns the absolute path of the database file.
	DBPath() string
}

// InitStorageUC is the use case executed on `vext init`.
type InitStorageUC struct {
	initialiser StorageInitialiser
}

func NewInitStorageUC(i StorageInitialiser) *InitStorageUC {
	return &InitStorageUC{initialiser: i}
}

// Execute prepares storage and restricts file permissions.
func (uc *InitStorageUC) Execute(ctx context.Context) error {
	if err := uc.initialiser.Init(ctx); err != nil {
		return fmt.Errorf("storage initialisation failed: %w", err)
	}

	// Restrict database file to owner only (0600).
	if err := os.Chmod(uc.initialiser.DBPath(), 0600); err != nil {
		return fmt.Errorf("could not set database permissions: %w", err)
	}

	return nil
}
