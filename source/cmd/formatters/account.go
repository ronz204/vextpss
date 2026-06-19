package formatters

import (
	"fmt"

	"vextpss/source/secrets"
)

func PrintAccount(name string, s secrets.AccountSecret) {
	fmt.Printf("Name:     %s\n", name)
	fmt.Printf("Username: %s\n", s.Username)
	fmt.Printf("Password: %s\n", s.Password)
}
