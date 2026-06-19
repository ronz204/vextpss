package shared

import (
	"vextpss/source/cmd/collectors"
	"vextpss/source/shared/cryptors"
)

// AppDeps holds the application-wide dependencies shared across all adapters.
// root.go constructs this once and passes it to every command.
type AppDeps struct {
	DBPath    string
	Enc       *cryptors.AESGCMEncryptor
	Collector *collectors.Collector
}
