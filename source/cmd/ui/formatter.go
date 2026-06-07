package ui

import (
	"fmt"

	"vextpss/source/core"
	"vextpss/source/core/secrets"
	"vextpss/source/shared"
)

// PrintSecret displays a decrypted secret to stdout in a human-readable format.
func PrintSecret(name string, payload core.SecretPayload) {
	switch p := payload.(type) {
	case *secrets.AccountSecret:
		fmt.Printf("Service:  %s\n", name)
		fmt.Printf("Username: %s\n", p.Username)
		fmt.Printf("Password: %s\n", p.Password)
		defer shared.Zero(p.Password)
	case *secrets.CreditSecret:
		printCreditSecret(name, p)
	default:
		fmt.Printf("Service: %s (unknown type)\n", name)
	}
}

func printCreditSecret(name string, p *secrets.CreditSecret) {
	fmt.Printf("Service:        %s\n", name)
	fmt.Printf("Card Number:    %s\n", formatCardNumber(p.Number))
	fmt.Printf("Security Code:  %s\n", p.SecurityCode)
	fmt.Printf("Expires:        %02d/%d\n", p.ExpirationMonth, p.ExpirationYear)
	fmt.Printf("PIN:            %s\n", p.Pin)
	defer shared.Zero(p.SecurityCode)
	defer shared.Zero(p.Pin)

	if p.BankUsername != "" {
		fmt.Printf("Bank Username:  %s\n", p.BankUsername)
	}
	if len(p.BankPassword) > 0 {
		fmt.Printf("Bank Password:  %s\n", p.BankPassword)
		defer shared.Zero(p.BankPassword)
	}
	if len(p.BankVirtualKey) > 0 {
		fmt.Printf("Virtual Key:    %s\n", p.BankVirtualKey)
		defer shared.Zero(p.BankVirtualKey)
	}
	if p.BankCellphone != "" {
		fmt.Printf("Cellphone:      %s\n", p.BankCellphone)
	}
	if p.CountryCode != "" {
		fmt.Printf("Country:        %s\n", p.CountryCode)
	}
}

// formatCardNumber inserts spaces every 4 digits for readability.
func formatCardNumber(number string) string {
	if len(number) != 16 {
		return number
	}
	return number[:4] + " " + number[4:8] + " " + number[8:12] + " " + number[12:]
}

// PrintSecretList displays a table of secret metadata.
func PrintSecretList(secrets []core.Secret) {
	if len(secrets) == 0 {
		fmt.Println("No secrets stored. Use `vext add <name>` to add one.")
		return
	}

	fmt.Printf("%-30s  %-12s  %s\n", "NAME", "TYPE", "CREATED")
	fmt.Printf("%-30s  %-12s  %s\n", "------------------------------", "------------", "-------------------")
	for _, s := range secrets {
		fmt.Printf("%-30s  %-12s  %s\n", s.Name, s.Type, s.CreatedAt.Format("2006-01-02 15:04:05"))
	}
}
