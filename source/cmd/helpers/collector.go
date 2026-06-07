package helpers

import (
	"fmt"

	"vextpss/source/core"
	"vextpss/source/core/secrets"
)

// SecretCollector knows how to interactively collect all fields for one secret type.
// It uses a Prompter for I/O, keeping collection logic separate from low-level terminal ops.
type SecretCollector interface {
	Collect(p Prompter) (core.SecretPayload, error)
}

// CollectorOptions carries optional parameters that some collectors may use.
type CollectorOptions struct {
	Generate   bool
	GenLength  int
	GenSymbols bool
}

// NewSecretCollector returns the collector for the given secret type.
// Adding a new secret type requires one new case here and a new collector file.
func NewSecretCollector(secretType string, opts CollectorOptions) (SecretCollector, error) {
	switch secretType {
	case secrets.TypeAccount:
		return &AccountCollector{opts: opts}, nil
	case secrets.TypeCredit:
		return &CreditCollector{}, nil
	default:
		return nil, core.NewDomainError(fmt.Sprintf("unknown secret type %q", secretType))
	}
}
