package commands

import (
	"fmt"
	"os"
	"path/filepath"

	"vextpss/source/pkg/database"

	"github.com/spf13/cobra"
)

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize the Vext environment",
	Long:  "Creates the config directory and database on first use. Safe to run multiple times.",
	RunE:  runInit,
}

func runInit(cmd *cobra.Command, args []string) error {
	dbPath, err := database.DBPath()
	if err != nil {
		return err
	}

	appDir := filepath.Dir(dbPath)

	// Create the config directory with restricted permissions.
	if err := os.MkdirAll(appDir, 0700); err != nil {
		return fmt.Errorf("could not create config directory: %w", err)
	}

	// Open (or create) the database file.
	db, err := database.Open(dbPath)
	if err != nil {
		return err
	}
	defer db.Close()

	// Apply the schema (idempotent — CREATE TABLE IF NOT EXISTS).
	if err := database.Migrate(db); err != nil {
		return err
	}

	// Restrict database file permissions to owner only.
	if err := os.Chmod(dbPath, 0600); err != nil {
		return fmt.Errorf("could not set database permissions: %w", err)
	}

	fmt.Printf("[✓] Vext initialized at %s\n", dbPath)
	return nil
}
