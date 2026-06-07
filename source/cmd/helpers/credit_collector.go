package helpers

import (
	"fmt"
	"strconv"
	"strings"

	"vextpss/source/core"
	"vextpss/source/core/secrets"
)

// CreditCollector collects all fields for a credit card secret.
// Required fields: card number, security code, expiration month/year, PIN.
// Optional fields: bank username, bank password, virtual key, cellphone, country code.
type CreditCollector struct{}

func (c *CreditCollector) Collect(p Prompter) (core.SecretPayload, error) {
	number, err := p.ReadLine("Card Number: ")
	if err != nil {
		return nil, fmt.Errorf("could not read card number: %w", err)
	}
	number = strings.ReplaceAll(number, " ", "")

	securityCode, err := p.ReadPassword("Security Code (CVV): ")
	if err != nil {
		return nil, fmt.Errorf("could not read security code: %w", err)
	}

	expMonthStr, err := p.ReadLine("Expiration Month (1-12): ")
	if err != nil {
		return nil, fmt.Errorf("could not read expiration month: %w", err)
	}
	expMonth, err := strconv.Atoi(strings.TrimSpace(expMonthStr))
	if err != nil || expMonth < 1 || expMonth > 12 {
		return nil, core.NewDomainError(fmt.Sprintf("invalid expiration month %q: must be a number between 1 and 12", expMonthStr))
	}

	expYearStr, err := p.ReadLine("Expiration Year (YYYY): ")
	if err != nil {
		return nil, fmt.Errorf("could not read expiration year: %w", err)
	}
	expYear, err := strconv.Atoi(strings.TrimSpace(expYearStr))
	if err != nil || expYear < 2000 || expYear > 2100 {
		return nil, core.NewDomainError(fmt.Sprintf("invalid expiration year %q", expYearStr))
	}

	pin, err := p.ReadPassword("PIN: ")
	if err != nil {
		return nil, fmt.Errorf("could not read PIN: %w", err)
	}

	fmt.Println("-- Optional fields (leave blank to skip) --")

	bankUsername, err := p.ReadLine("Bank Username: ")
	if err != nil {
		return nil, fmt.Errorf("could not read bank username: %w", err)
	}

	var bankPassword []byte
	if bankUsername != "" {
		bankPassword, err = p.ReadPassword("Bank Password: ")
		if err != nil {
			return nil, fmt.Errorf("could not read bank password: %w", err)
		}
	}

	bankVirtualKey, err := p.ReadPassword("Bank Virtual Key: ")
	if err != nil {
		return nil, fmt.Errorf("could not read bank virtual key: %w", err)
	}

	cellphone, err := p.ReadLine("Cellphone: ")
	if err != nil {
		return nil, fmt.Errorf("could not read cellphone: %w", err)
	}

	countryCode, err := p.ReadLine("Country Code: ")
	if err != nil {
		return nil, fmt.Errorf("could not read country code: %w", err)
	}

	return &secrets.CreditSecret{
		Number:          number,
		SecurityCode:    securityCode,
		ExpirationMonth: expMonth,
		ExpirationYear:  expYear,
		Pin:             pin,
		BankUsername:    bankUsername,
		BankPassword:    bankPassword,
		BankVirtualKey:  bankVirtualKey,
		BankCellphone:   cellphone,
		CountryCode:     countryCode,
	}, nil
}
