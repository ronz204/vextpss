package ui

import (
	"fmt"
	"vextpss/source/pkg/domain"
	"vextpss/source/pkg/shared"
)

// PrintSecret displays a decrypted secret to stdout in a human-readable format.
func PrintSecret(name string, payload domain.SecretPayload) {
	switch p := payload.(type) {
	case *domain.AccountSecret:
		fmt.Printf("Service:  %s\n", name)
		fmt.Printf("Username: %s\n", p.Username)
		fmt.Printf("Password: %s\n", p.Password)
		defer shared.Zero(p.Password)
	default:
		fmt.Printf("Service: %s (unknown type)\n", name)
	}
}

// PrintSecretList displays a table of secret metadata.
func PrintSecretList(secrets []domain.Secret) {
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
