package collectors

import (
	"fmt"

	"vextpss/source/secrets"
)

func Payload(p *Prompter, secretType string) ([]byte, error) {
	switch secretType {
	case secrets.TypeAccount:
		return collectAccount(p)
	case secrets.TypeFinance:
		return collectFinance(p)
	default:
		return nil, fmt.Errorf("unknown secret type: %q", secretType)
	}
}

func Master(p *Prompter) ([]byte, error) {
	return p.ReadSecret("Master password")
}
