package storage

import (
	"fmt"
	"os"
	"path/filepath"
)

const appName = "vext"

func DBPath() (string, error) {
	dir, err := os.UserConfigDir()
	if err != nil {
		return "", fmt.Errorf("could not determine config directory: %w", err)
	}
	return filepath.Join(dir, appName, appName+".db"), nil
}
