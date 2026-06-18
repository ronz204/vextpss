package core

// Collector is the input port for gathering user-provided data needed by use cases.
// The cmd/collectors package provides the terminal implementation.
type Collector interface {
	CollectPayload(secretType string) ([]byte, error)
	CollectMaster() ([]byte, error)
	Confirm(prompt string) (bool, error)
}
