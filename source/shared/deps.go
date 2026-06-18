package shared

import "vextpss/source/core"

// AppDeps holds the application-wide dependencies shared across all adapters.
// root.go constructs this once and passes it to every command.
type AppDeps struct {
	DBPath    string
	Enc       core.Encryptor
	Collector core.Collector
}
