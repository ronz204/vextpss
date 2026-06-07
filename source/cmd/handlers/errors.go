package handlers

import (
	"fmt"

	"vextpss/source/core"
)

// printErr prints a user-facing error message and returns true when err is non-nil.
// notFoundMsg is printed when err is ErrSecretNotFound; pass "" to fall back to the domain message.
func printErr(err error, notFoundMsg string) bool {
	if err == nil {
		return false
	}
	if notFoundMsg != "" && core.IsNotFound(err) {
		fmt.Printf("[X] %s\n", notFoundMsg)
		return true
	}
	if core.IsDomainError(err) {
		fmt.Printf("[X] %s\n", err)
		return true
	}
	fmt.Println("[X] An unexpected error occurred. Please try again.")
	return true
}
