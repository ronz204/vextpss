package handlers

import (
	"fmt"
	"os"
)

// guardInit returns false and prints an actionable error if the vault has not been initialised.
// Use as: if !guardInit(h.uc != nil) { return nil }
func guardInit(initialised bool) bool {
	if !initialised {
		fmt.Fprintln(os.Stderr, "[X] Vault not initialised. Run 'vext init' first.")
	}
	return initialised
}
