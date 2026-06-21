package collectors

import (
	"encoding/json"
	"fmt"

	"vextpss/source/secrets"
	"vextpss/source/shared/memory"
)

func collectFinance(p *Prompter) ([]byte, error) {
	cardNumber, err := p.ReadLine("Card Number")
	if err != nil {
		return nil, err
	}

	cardPin, err := p.ReadSecret("Card PIN")
	defer memory.Cleaner(cardPin)
	if err != nil {
		return nil, err
	}

	securityCode, err := p.ReadSecret("Security code (CVV)")
	defer memory.Cleaner(securityCode)
	if err != nil {
		return nil, err
	}

	expirationMonth, err := p.ReadInteger("Expiration month")
	if err != nil {
		return nil, err
	}

	expirationYear, err := p.ReadInteger("Expiration year")
	if err != nil {
		return nil, err
	}

	bankUsername, err := p.ReadLine("App Username")
	if err != nil {
		return nil, err
	}

	bankPassword, err := p.ReadSecret("App Password")
	defer memory.Cleaner(bankPassword)
	if err != nil {
		return nil, err
	}

	bankVirtualKey, err := p.ReadSecret("Security Token")
	defer memory.Cleaner(bankVirtualKey)
	if err != nil {
		return nil, err
	}

	bankCellphone, err := p.ReadLine("Cellphone")
	if err != nil {
		return nil, err
	}

	payload, err := json.Marshal(secrets.FinanceSecret{
		CardNumber:      cardNumber,
		CardPin:         cardPin,
		SecurityCode:    securityCode,
		ExpirationMonth: expirationMonth,
		ExpirationYear:  expirationYear,
		BankUsername:    bankUsername,
		BankPassword:    bankPassword,
		BankVirtualKey:  bankVirtualKey,
		BankCellphone:   bankCellphone,
	})
	if err != nil {
		return nil, fmt.Errorf("serialize finance: %w", err)
	}

	return payload, nil
}
