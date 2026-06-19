package cmd

import (
	"vextpss/source/cmd/collectors"
	"vextpss/source/funcs"
	"vextpss/source/shared/cryptors"
)

// AppDeps holds the two behavioral dependencies shared across all adapters.
type AppDeps struct {
	Encryptor funcs.Encryptor
	Collector funcs.Collector
}

// Build constructs AppDeps with all concrete implementations wired.
func Build() AppDeps {
	return AppDeps{
		Encryptor: cryptors.NewAESGCMEncryptor(cryptors.DefaultConfig()),
		Collector: collectors.NewCollector(collectors.NewPrompter()),
	}
}
