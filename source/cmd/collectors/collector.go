package collectors

import (
	"fmt"

	"vextpss/source/secrets"
)

type Collector struct{ p *Prompter }

func NewCollector(p *Prompter) *Collector {
	return &Collector{p: p}
}

func (c *Collector) Payload(secretType string) ([]byte, error) {
	switch secretType {
	case secrets.TypeAccount:
		return collectAccount(c.p)
	case secrets.TypeFinance:
		return collectFinance(c.p)
	default:
		return nil, fmt.Errorf("unknown secret type: %q", secretType)
	}
}

func (c *Collector) Master() ([]byte, error) {
	return c.p.ReadSecret("Master password")
}

func (c *Collector) Confirm(prompt string) (bool, error) {
	return c.p.Confirm(prompt)
}
