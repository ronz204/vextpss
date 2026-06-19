package adapters

import (
	"vextpss/source/cmd/collectors"
	"vextpss/source/funcs"
	"vextpss/source/shared/cryptors"
	"vextpss/source/shared/storage"
)

type Deps struct {
	DBPath    string
	Encryptor funcs.Encryptor
	Collector funcs.Collector
}

func BuildDeps() Deps {
	return Deps{
		DBPath:    storage.DBPath(),
		Encryptor: cryptors.NewAESGCMEncryptor(cryptors.DefaultConfig()),
		Collector: collectors.NewCollector(collectors.NewPrompter()),
	}
}
