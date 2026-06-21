package storage

import (
	"fmt"
	"os"
	"path/filepath"
)

const appName = "vext"

func DBPath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("could not determine home directory: %w", err)
	}
	return filepath.Join(home, ".config", appName, appName+".db"), nil
}
