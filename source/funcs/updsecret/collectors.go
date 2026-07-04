package updsecret

import (
	"encoding/json"
	"fmt"

	"vextpss/source/secrets/core"
	"vextpss/source/secrets/moon/accounts"
	"vextpss/source/secrets/moon/finances"
	"vextpss/source/shared/memory"
	"vextpss/source/shared/terminal"
)

func collect(secretType string, p *terminal.Prompter) (plaintext, master []byte, err error) {
	fn, ok := collectors[secretType]
	if !ok {
		return nil, nil, fmt.Errorf("unknown secret type: %s", secretType)
	}

	payload, err := fn(p)
	if err != nil {
		return nil, nil, err
	}

	master, err = p.ReadSecret("Master password")
	if err != nil {
		return nil, nil, err
	}

	plaintext, err = json.Marshal(payload)
	if err != nil {
		memory.Cleaner(master)
		return nil, nil, err
	}

	return plaintext, master, nil
}

func collectAccount(p *terminal.Prompter) (core.Payload, error) {
	username, err := p.ReadLine("Username")
	if err != nil {
		return nil, err
	}
	password, err := p.ReadSecret("Password")
	if err != nil {
		return nil, err
	}
	return accounts.NewAccount(username, password)
}

func collectFinance(p *terminal.Prompter) (core.Payload, error) {
	cardNumber, err := p.ReadLine("Card number")
	if err != nil {
		return nil, err
	}
	cardPin, err := p.ReadSecret("Card PIN")
	if err != nil {
		return nil, err
	}
	cardSecCode, err := p.ReadSecret("Security code")
	if err != nil {
		return nil, err
	}
	expMonth, err := p.ReadInteger("Expiration month (1-12)")
	if err != nil {
		return nil, err
	}
	expYear, err := p.ReadInteger("Expiration year")
	if err != nil {
		return nil, err
	}

	card, err := finances.NewCard(cardNumber, cardPin, cardSecCode, expMonth, expYear)
	if err != nil {
		return nil, err
	}

	mobileUsername, err := p.ReadLine("Mobile banking username")
	if err != nil {
		return nil, err
	}
	mobilePassword, err := p.ReadSecret("Mobile banking password")
	if err != nil {
		return nil, err
	}
	mobileVirtKey, err := p.ReadSecret("Virtual key")
	if err != nil {
		return nil, err
	}
	mobileCellphone, err := p.ReadLine("Cellphone")
	if err != nil {
		return nil, err
	}

	mobile, err := finances.NewMobile(mobileUsername, mobilePassword, mobileVirtKey, mobileCellphone)
	if err != nil {
		return nil, err
	}

	return finances.NewFinance(card, mobile), nil
}
