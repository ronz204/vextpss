package secrets

import (
	"fmt"

	"vextpss/source/core"
)

// CreditSecret is the payload for secrets of type "credit".
// Sensitive fields are []byte so they can be zeroed from memory after use.
type CreditSecret struct {
	Number          string `json:"number"`
	SecurityCode    []byte `json:"security_code"`
	ExpirationMonth int    `json:"expiration_month"`
	ExpirationYear  int    `json:"expiration_year"`
	Pin             []byte `json:"pin"`
	BankUsername    string `json:"bank_username"`
	BankPassword    []byte `json:"bank_password"`
	BankVirtualKey  []byte `json:"virtual_key"`
	BankCellphone   string `json:"cellphone"`
	CountryCode     string `json:"country_code"`
}

func (c *CreditSecret) GetType() string { return "credit" }

func (c *CreditSecret) Validate() error {
	if c.Number == "" {
		return core.NewDomainError("card number is required")
	}
	if len(c.SecurityCode) == 0 {
		return core.NewDomainError("security code is required")
	}
	if c.ExpirationMonth < 1 || c.ExpirationMonth > 12 {
		return core.NewDomainError(fmt.Sprintf("expiration month must be between 1 and 12, got %d", c.ExpirationMonth))
	}
	if c.ExpirationYear < 2000 || c.ExpirationYear > 2100 {
		return core.NewDomainError(fmt.Sprintf("expiration year %d is out of range", c.ExpirationYear))
	}
	if len(c.Pin) == 0 {
		return core.NewDomainError("PIN is required")
	}
	return nil
}
