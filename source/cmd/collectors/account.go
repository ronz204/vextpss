package collectors

import (
	"encoding/json"
	"fmt"
	"strings"

	"vextpss/source/secrets"
	"vextpss/source/shared/memory"
)

func collectAccount(p *Prompter) ([]byte, error) {
	username, err := p.ReadLine("Username")
	if err != nil {
		return nil, err
	}

	if strings.TrimSpace(username) == "" {
		return nil, fmt.Errorf("%w: username is required", secrets.ErrInvalidInput)
	}

	password, err := p.ReadSecret("Password")
	defer memory.Cleaner(password)
	if err != nil {
		return nil, err
	}

	if len(password) == 0 {
		return nil, fmt.Errorf("%w: password is required", secrets.ErrInvalidInput)
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
