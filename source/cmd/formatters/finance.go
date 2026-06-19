package formatters

import (
	"fmt"

	"vextpss/source/secrets"
)

// ======================================
// PrintFinance prints all fields of a finance secret.
// Optional bank fields are skipped when empty.
// ======================================
func PrintFinance(name string, s secrets.FinanceSecret) {
	fmt.Printf("Name:        %s\n", name)
	fmt.Printf("Card number: %s\n", s.CardNumber)
	fmt.Printf("CVV:         %s\n", s.SecurityCode)
	fmt.Printf("Expiration:  %02d/%d\n", s.ExpirationMonth, s.ExpirationYear)
	fmt.Printf("Card PIN:    %s\n", s.CardPin)

	if s.BankUsername != "" {
		fmt.Printf("Bank user:   %s\n", s.BankUsername)
	}
	if len(s.BankPassword) > 0 {
		fmt.Printf("Bank pass:   %s\n", s.BankPassword)
	}
	if len(s.BankVirtualKey) > 0 {
		fmt.Printf("Virtual key: %s\n", s.BankVirtualKey)
	}
	if s.BankCellphone != "" {
		fmt.Printf("Cellphone:   %s\n", s.BankCellphone)
	}
}
