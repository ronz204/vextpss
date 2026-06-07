package ui

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

// collectorFactory builds a SecretCollector for a given type and options.
type collectorFactory func(opts CollectorOptions) SecretCollector

// collectorRegistry maps each secret type to its factory.
// Adding a new secret type requires one new entry here.
var collectorRegistry = map[string]collectorFactory{
	secrets.TypeAccount: func(opts CollectorOptions) SecretCollector { return &AccountCollector{opts: opts} },
	secrets.TypeCredit:  func(_ CollectorOptions) SecretCollector { return &CreditCollector{} },
}

// NewSecretCollector returns the collector for the given secret type.
func NewSecretCollector(secretType string, opts CollectorOptions) (SecretCollector, error) {
	factory, ok := collectorRegistry[secretType]
	if !ok {
		return nil, core.NewDomainError(fmt.Sprintf("unknown secret type %q", secretType))
	}
	return factory(opts), nil
}
