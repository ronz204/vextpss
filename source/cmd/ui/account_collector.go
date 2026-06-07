package ui

import (
	"fmt"

	"vextpss/source/core"
	"vextpss/source/core/secrets"
	"vextpss/source/shared"
)

// AccountCollector collects username and password for an account secret.
type AccountCollector struct {
	opts CollectorOptions
}

func (c *AccountCollector) Collect(p Prompter) (core.SecretPayload, error) {
	username, err := p.ReadLine("Username: ")
	if err != nil {
		return nil, fmt.Errorf("could not read username: %w", err)
	}

	var password []byte
	if c.opts.Generate {
		pw, err := shared.GeneratePassword(c.opts.GenLength, c.opts.GenSymbols)
		if err != nil {
			return nil, fmt.Errorf("failed to generate password: %w", err)
		}
		password = []byte(pw)
		fmt.Printf("Generated password: %s\n", pw)
	} else {
		password, err = p.ReadPassword("Password: ")
		if err != nil {
			return nil, fmt.Errorf("could not read password: %w", err)
		}
	}

	return &secrets.AccountSecret{
		Username: username,
		Password: password,
	}, nil
}
