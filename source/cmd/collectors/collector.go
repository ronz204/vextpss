package collectors

import (
	"fmt"
	"strings"

	"vextpss/source/secrets"
	"vextpss/source/shared/sentinel"
)

type Collector struct {
	p *Prompter
}

func NewCollector(p *Prompter) *Collector {
	return &Collector{p: p}
}

// ======================================
// Payload dispatches to the right collector based on secret type.
// Add a new case here when a new type is introduced.
// ======================================
func (c *Collector) Payload(secretType string) ([]byte, error) {
	switch secretType {
	case secrets.TypeAccount:
		return collectAccount(c.p)
	case secrets.TypeFinance:
		return collectFinance(c.p)
	default:
		return nil, fmt.Errorf("%w: type %q is not yet supported", sentinel.ErrInvalidInput, secretType)
	}
}

// ======================================
// Master prompts for the master password without echoing it.
// ======================================
func (c *Collector) Master() ([]byte, error) {
	return c.p.ReadSecret("Master password")
}

// ======================================
// Confirm asks the user a yes/no question and returns true only for "y".
// ======================================
func (c *Collector) Confirm(prompt string) (bool, error) {
	answer, err := c.p.ReadLine(fmt.Sprintf("%s (y/N)", prompt))
	if err != nil {
		return false, err
	}
	return strings.ToLower(answer) == "y", nil
}
