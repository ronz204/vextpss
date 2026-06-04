package cmd

import "github.com/spf13/cobra"

// NewRootCmd builds the root Cobra command. Sub-commands are registered in main.go
// so dependency injection can be performed before the command tree is assembled.
func NewRootCmd() *cobra.Command {
	return &cobra.Command{
		Use:           "vext",
		Short:         "A local-first CLI password manager",
		Long:          "Vext stores and retrieves credentials securely using AES-256-GCM encryption and Argon2id key derivation. All data lives on your machine — no cloud, no accounts.",
		SilenceErrors: true,
		SilenceUsage:  true,
	}
}
