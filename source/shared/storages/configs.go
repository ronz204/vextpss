package storages

import (
	"os"
	"path/filepath"
)

const (
	dbFileName          = "vext.db"
	DefaultActiveSpace  = "default"
)

func DefaultDBPath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".config", "vext", dbFileName), nil
}
