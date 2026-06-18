package collectors

import (
	"encoding/json"
	"fmt"

	"vextpss/source/secrets"
	"vextpss/source/shared/passgen"
)

// ================================
// collectAccount prompts for username and password, then marshals them into a JSON payload.
// rawPassword is zeroed immediately after marshaling.
// ================================
func collectAccount(p Prompter) ([]byte, error) {
	username, err := p.ReadLine("Username")
	if err != nil {
		return nil, err
	}

	rawPassword, err := p.ReadSecret("Password")
	defer passgen.Cleaner(rawPassword)
	if err != nil {
		return nil, err
	}

	payload, err := json.Marshal(secrets.AccountSecret{
		Username: username,
		Password: rawPassword,
	})
	if err != nil {
		return nil, fmt.Errorf("serialization failed: %w", err)
	}

	return payload, nil
}
