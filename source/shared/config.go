package shared

import (
	"os"
	"path/filepath"
)

// AppName is the canonical application identifier used for directories and the binary.
const AppName = "vext"

// DBPath returns the platform-specific path to the application's SQLite database.
func DBPath() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".config", AppName, AppName+".db")
}
