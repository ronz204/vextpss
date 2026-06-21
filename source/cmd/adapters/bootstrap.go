package adapters

import (
	"fmt"

	"vextpss/source/cmd/collectors"
	"vextpss/source/secrets"
	"vextpss/source/shared/cryptors"
	"vextpss/source/shared/storage"
)

type App struct {
	DBPath     string
	Repository secrets.Repository
	Encryptor  secrets.Encryptor
	Collector  *collectors.Collector
}

func Build() (App, func(), error) {
	dbPath, err := storage.DBPath()
	if err != nil {
		return App{}, nil, err
	}

	db, err := storage.Open(dbPath)
	if err != nil {
		return App{}, nil, fmt.Errorf("open database: %w", err)
	}

	return App{
		DBPath:     dbPath,
		Repository: storage.NewSecretRepository(db),
		Encryptor:  cryptors.NewAESGCMEncryptor(cryptors.DefaultConfig()),
		Collector:  collectors.NewCollector(collectors.NewPrompter()),
	}, func() { _ = storage.Close(db) }, nil
}
