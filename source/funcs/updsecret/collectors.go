package updsecret

import (
	"encoding/json"
	"fmt"

	"vextpss/source/secrets/core"
	"vextpss/source/secrets/moon/accounts"
	"vextpss/source/secrets/moon/finances"
	"vextpss/source/shared/terminal"
)

type payloadCollector func(*terminal.Prompter, []byte) (core.Payload, error)

var collectors = map[string]payloadCollector{
	core.TypeAccount: collectAccount,
	core.TypeFinance: collectFinance,
}

func collect(secretType string, current []byte, p *terminal.Prompter) ([]byte, error) {
	fn, ok := collectors[secretType]
	if !ok {
		return nil, fmt.Errorf("unknown secret type: %s", secretType)
	}

	payload, err := fn(p, current)
	if err != nil {
		return nil, err
	}

	return json.Marshal(payload)
}

func collectAccount(p *terminal.Prompter, cur []byte) (core.Payload, error) {
	var current accounts.Account
	if err := json.Unmarshal(cur, &current); err != nil {
		return nil, err
	}

	username, err := p.ReadLineOrKeep("Username", current.Username)
	if err != nil {
		return nil, err
	}
	password, err := p.ReadSecretOrKeep("Password", current.Password)
	if err != nil {
		return nil, err
	}
	return accounts.NewAccount(username, password)
}

func collectFinance(p *terminal.Prompter, cur []byte) (core.Payload, error) {
	var current finances.Finance
	if err := json.Unmarshal(cur, &current); err != nil {
		return nil, err
	}

	cardNumber, err := p.ReadLineOrKeep("Card number", current.Card.Number)
	if err != nil {
		return nil, err
	}
	cardPin, err := p.ReadSecretOrKeep("Card PIN", current.Card.Pin)
	if err != nil {
		return nil, err
	}
	cardSecCode, err := p.ReadSecretOrKeep("Security code", current.Card.SecurityCode)
	if err != nil {
		return nil, err
	}
	expMonth, err := p.ReadIntegerOrKeep("Expiration month (1-12)", current.Card.ExpirationMonth)
	if err != nil {
		return nil, err
	}
	expYear, err := p.ReadIntegerOrKeep("Expiration year", current.Card.ExpirationYear)
	if err != nil {
		return nil, err
	}

	card, err := finances.NewCard(cardNumber, cardPin, cardSecCode, expMonth, expYear)
	if err != nil {
		return nil, err
	}

	mobileUsername, err := p.ReadLineOrKeep("Mobile banking username", current.Mobile.Username)
	if err != nil {
		return nil, err
	}
	mobilePassword, err := p.ReadSecretOrKeep("Mobile banking password", current.Mobile.Password)
	if err != nil {
		return nil, err
	}
	mobileVirtKey, err := p.ReadSecretOrKeep("Virtual key", current.Mobile.VirtualKey)
	if err != nil {
		return nil, err
	}
	mobileCellphone, err := p.ReadLineOrKeep("Cellphone", current.Mobile.Cellphone)
	if err != nil {
		return nil, err
	}

	mobile, err := finances.NewMobile(mobileUsername, mobilePassword, mobileVirtKey, mobileCellphone)
	if err != nil {
		return nil, err
	}

	return finances.NewFinance(card, mobile), nil
}
