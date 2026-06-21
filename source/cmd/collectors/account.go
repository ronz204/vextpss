package collectors

import (
	"encoding/json"
	"fmt"

	"vextpss/source/secrets"
	"vextpss/source/shared/memory"
)

func collectAccount(p *Prompter) ([]byte, error) {
	username, err := p.ReadLine("Username")
	if err != nil {
		return nil, err
	}

	password, err := p.ReadSecret("Password")
	defer memory.Cleaner(password)
	if err != nil {
		return nil, err
	}

	payload, err := json.Marshal(secrets.AccountSecret{
		Username: username,
		Password: password,
	})
	if err != nil {
		return nil, fmt.Errorf("serialize account: %w", err)
	}

	return payload, nil
}
