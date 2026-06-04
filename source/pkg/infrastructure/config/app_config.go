package config

import (
	"fmt"
	"os"
	"path/filepath"
)

const (
	appDirName = "vext"
	dbFileName = "vext.db"
)

// AppConfig holds all runtime configuration for the application.
type AppConfig struct {
	DBPath string
	AppDir string
}

// Load resolves OS-appropriate paths and returns the application config.
// Linux/macOS: ~/.config/vext/vext.db
// Windows:     %AppData%\vext\vext.db
func Load() (*AppConfig, error) {
	configDir, err := os.UserConfigDir()
	if err != nil {
		return nil, fmt.Errorf("could not determine config directory: %w", err)
	}
	appDir := filepath.Join(configDir, appDirName)
	return &AppConfig{
		DBPath: filepath.Join(appDir, dbFileName),
		AppDir: appDir,
	}, nil
}
