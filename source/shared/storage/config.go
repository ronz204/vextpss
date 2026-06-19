package storage

import (
	"os"
	"path/filepath"
)

const appName = "vext"

// DBPath returns the platform-specific path to the application's SQLite database.
func DBPath() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".config", appName, appName+".db")
}
