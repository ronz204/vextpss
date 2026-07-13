package storages

import (
	"context"
	"os"
	"path/filepath"

	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"vextpss/source/secrets/core"
)

const dbFileName = "vext.db"

func DefaultDBPath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".config", "vext", dbFileName), nil
}

func Init(dbPath string) (*gorm.DB, core.State, error) {
	db, err := openDB(dbPath)
	if err != nil {
		return nil, core.State{}, err
	}
	if err := migrate(db); err != nil {
		return nil, core.State{}, err
	}
	state, err := seed(db)
	if err != nil {
		return nil, core.State{}, err
	}
	return db, state, nil
}

func openDB(dbPath string) (*gorm.DB, error) {
	if err := os.MkdirAll(filepath.Dir(dbPath), 0700); err != nil {
		return nil, err
	}
	return gorm.Open(sqlite.Open(dbPath), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
}

func migrate(db *gorm.DB) error {
	return db.AutoMigrate(&SpaceRecord{}, &MetaRecord{}, &SecretRecord{})
}

func seed(db *gorm.DB) (core.State, error) {
	ctx := context.Background()
	if err := db.FirstOrCreate(&SpaceRecord{}, SpaceRecord{Name: core.DefaultActiveSpace}).Error; err != nil {
		return core.State{}, err
	}
	stateRepo := NewState(db)
	state, err := stateRepo.Load(ctx)
	if err != nil {
		return core.State{}, err
	}
	return state, stateRepo.Save(ctx, state)
}
