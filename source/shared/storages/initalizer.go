package storages

import (
	"os"
	"path/filepath"

	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func Init(dbPath string) (*gorm.DB, error) {
	if err := os.MkdirAll(filepath.Dir(dbPath), 0700); err != nil {
		return nil, err
	}
	db, err := gorm.Open(sqlite.Open(dbPath), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		return nil, err
	}
	if err := db.AutoMigrate(&SpaceRecord{}, &MetaRecord{}, &SecretRecord{}); err != nil {
		return nil, err
	}
	if err := db.FirstOrCreate(&SpaceRecord{}, SpaceRecord{Name: "default"}).Error; err != nil {
		return nil, err
	}
	if err := db.Where(MetaRecord{Key: activeSpaceKey}).
		FirstOrCreate(&MetaRecord{Key: activeSpaceKey, Value: "default"}).Error; err != nil {
		return nil, err
	}
	return db, nil
}
