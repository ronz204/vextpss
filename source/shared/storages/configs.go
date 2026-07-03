package storage

import (
	"os"
	"path/filepath"
)

const dbFileName = "vext.db"

func DefaultDBPath() (string, error) {
	dir, err := os.UserConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "vext", dbFileName), nil
}
