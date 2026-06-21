package storage

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
)

// SetupDB creates the config directory and applies schema migrations.
// Opens and closes its own connection — safe to call alongside an existing one.
func SetupDB(ctx context.Context, dbPath string) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(dbPath), 0700); err != nil {
		return fmt.Errorf("could not create config directory: %w", err)
	}
	db, err := Open(dbPath)
	if err != nil {
		return err
	}
	defer Close(db)
	return Migrate(db)
}
