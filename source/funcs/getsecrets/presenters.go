package getsecrets

import (
	"encoding/json"
	"fmt"

	"vextpss/source/secrets"
	"vextpss/source/shared/payloads/accounts"
	"vextpss/source/shared/payloads/finances"
	"vextpss/source/shared/terminal"
)

type payloadDisplay func(plaintext []byte) error

var displayers = map[string]payloadDisplay{
	secrets.TypeAccount: displayAccount,
	secrets.TypeFinance: displayFinance,
}

func display(secretType string, plaintext []byte) error {
	fn, ok := displayers[secretType]
	if !ok {
		return fmt.Errorf("unknown secret type %q", secretType)
	}
	return fn(plaintext)
}

func displayAccount(plaintext []byte) error {
	var a accounts.Account
	if err := json.Unmarshal(plaintext, &a); err != nil {
		return secrets.ErrCorruptPayload
	}
	terminal.PrintAccount(a)
	return nil
}

func displayFinance(plaintext []byte) error {
	var f finances.Finance
	if err := json.Unmarshal(plaintext, &f); err != nil {
		return secrets.ErrCorruptPayload
	}
	terminal.PrintFinance(f)
	return nil
}
