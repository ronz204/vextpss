package formatters

import (
	"encoding/json"
	"fmt"

	"vextpss/source/secrets"
)

func PrintSecret(name, secretType string, payload []byte) {
	fmt.Printf("Name: %s\n", name)
	fmt.Printf("Type: %s\n", secretType)

	switch secretType {
	case secrets.TypeAccount:
		printAccount(payload)
	case secrets.TypeFinance:
		printFinance(payload)
	default:
		fmt.Printf("Payload: %s\n", string(payload))
	}
}

func printAccount(payload []byte) {
	var s secrets.AccountSecret
	if err := json.Unmarshal(payload, &s); err != nil {
		fmt.Printf("(could not parse payload: %v)\n", err)
		return
	}

	fmt.Printf("Username: %s\n", s.Username)
	fmt.Printf("Password: %s\n", s.Password)
}

func printFinance(payload []byte) {
	var f secrets.FinanceSecret
	if err := json.Unmarshal(payload, &f); err != nil {
		fmt.Printf("(could not parse payload: %v)\n", err)
		return
	}

	fmt.Printf("Card number:      %s\n", f.CardNumber)
	fmt.Printf("Card PIN:         %s\n", f.CardPin)
	fmt.Printf("Security code:    %s\n", f.SecurityCode)
	fmt.Printf("Expiration:       %02d/%d\n", f.ExpirationMonth, f.ExpirationYear)
	fmt.Printf("Bank username:    %s\n", f.BankUsername)
	fmt.Printf("Bank password:    %s\n", f.BankPassword)
	fmt.Printf("Bank virtual key: %s\n", f.BankVirtualKey)
	fmt.Printf("Bank cellphone:   %s\n", f.BankCellphone)
}
