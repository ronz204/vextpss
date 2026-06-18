package collectors

import (
	"encoding/json"
	"fmt"
	"strconv"

	"vextpss/source/secrets"
	"vextpss/source/shared/passgen"
	"vextpss/source/shared/sentinel"
)

// ================================
// collectFinance prompts for card and bank credentials, then marshals them into a JSON payload.
// Sensitive byte slices are zeroed immediately after marshaling.
// Bank fields are optional — the user may leave them blank.
// ================================
func collectFinance(p Prompter) ([]byte, error) {
	cardNumber, err := p.ReadLine("Card number")
	if err != nil {
		return nil, err
	}

	securityCode, err := p.ReadSecret("CVV")
	defer passgen.Cleaner(securityCode)
	if err != nil {
		return nil, err
	}

	cardPin, err := p.ReadSecret("Card PIN")
	defer passgen.Cleaner(cardPin)
	if err != nil {
		return nil, err
	}

	expMonthStr, err := p.ReadLine("Expiry month (MM)")
	if err != nil {
		return nil, err
	}
	expMonth, err := strconv.Atoi(expMonthStr)
	if err != nil {
		return nil, fmt.Errorf("%w: expiry month must be a number", sentinel.ErrInvalidInput)
	}

	expYearStr, err := p.ReadLine("Expiry year (YYYY)")
	if err != nil {
		return nil, err
	}
	expYear, err := strconv.Atoi(expYearStr)
	if err != nil {
		return nil, fmt.Errorf("%w: expiry year must be a number", sentinel.ErrInvalidInput)
	}

	bankUsername, _ := p.ReadLine("Bank username")

	bankPassword, err := p.ReadSecret("Bank password")
	defer passgen.Cleaner(bankPassword)
	if err != nil {
		return nil, err
	}

	bankVirtualKey, err := p.ReadSecret("Bank virtual key")
	defer passgen.Cleaner(bankVirtualKey)
	if err != nil {
		return nil, err
	}

	bankCellphone, _ := p.ReadLine("Bank cellphone")

	payload, err := json.Marshal(secrets.FinanceSecret{
		CardNumber:      cardNumber,
		SecurityCode:    securityCode,
		CardPin:         cardPin,
		ExpirationMonth: expMonth,
		ExpirationYear:  expYear,
		BankUsername:    bankUsername,
		BankPassword:    bankPassword,
		BankVirtualKey:  bankVirtualKey,
		BankCellphone:   bankCellphone,
	})
	if err != nil {
		return nil, fmt.Errorf("serialization failed: %w", err)
	}

	return payload, nil
}
