package database

import (
	"fmt"
	"os"
	"path/filepath"
)

const appDirName = "vext"
const dbFileName = "vext.db"

// DBPath returns the OS-appropriate path for the Vext database file.
// Linux/macOS: ~/.config/vext/vext.db
// Windows:     %AppData%\vext\vext.db
func DBPath() (string, error) {
	configDir, err := os.UserConfigDir()
	if err != nil {
		return "", fmt.Errorf("could not determine config directory: %w", err)
	}
	return filepath.Join(configDir, appDirName, dbFileName), nil
}

// AppDir returns the OS-appropriate directory for Vext config files.
func AppDir() (string, error) {
	configDir, err := os.UserConfigDir()
	if err != nil {
		return "", fmt.Errorf("could not determine config directory: %w", err)
	}
	return filepath.Join(configDir, appDirName), nil
}
