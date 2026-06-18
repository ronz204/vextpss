package collectors

import (
	"fmt"
	"strings"

	"vextpss/source/secrets"
	"vextpss/source/shared/sentinel"
)

// Collector is the contract for gathering all inputs needed by an adapter.
// Adapters depend only on this interface — Prompter is an implementation detail.
type Collector interface {
	CollectPayload(secretType string) ([]byte, error)
	CollectMaster() ([]byte, error)
	Confirm(prompt string) (bool, error)
}

// TerminalCollector implements Collector using a Prompter for raw terminal input.
type TerminalCollector struct {
	p Prompter
}

func NewTerminalCollector(p Prompter) *TerminalCollector {
	return &TerminalCollector{p: p}
}

// ================================
// CollectPayload dispatches to the right collector based on secret type.
// Add a new case here when a new type is introduced.
// ================================
func (c *TerminalCollector) CollectPayload(secretType string) ([]byte, error) {
	switch secretType {
	case secrets.TypeAccount:
		return collectAccount(c.p)
	case secrets.TypeFinance:
		return collectFinance(c.p)
	default:
		return nil, fmt.Errorf("%w: type %q is not yet supported", sentinel.ErrInvalidInput, secretType)
	}
}

// ================================
// CollectMaster prompts for the master password without echoing it.
// ================================
func (c *TerminalCollector) CollectMaster() ([]byte, error) {
	return c.p.ReadSecret("Master password")
}

// ================================
// Confirm asks the user a yes/no question and returns true only for "y".
// ================================
func (c *TerminalCollector) Confirm(prompt string) (bool, error) {
	answer, err := c.p.ReadLine(fmt.Sprintf("%s (y/N)", prompt))
	if err != nil {
		return false, err
	}
	return strings.ToLower(answer) == "y", nil
}
