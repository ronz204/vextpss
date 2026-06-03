package commands

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var RootCmd = &cobra.Command{
	Use:           "vext",
	Short:         "A local-first CLI password manager",
	Long:          "Vext stores and retrieves credentials securely using AES-256-GCM encryption and Argon2id key derivation. All data lives on your machine — no cloud, no accounts.",
	SilenceErrors: true,
	SilenceUsage:  true,
}

func init() {
	RootCmd.AddCommand(initCmd)
	RootCmd.AddCommand(addCmd)
	RootCmd.AddCommand(getCmd)
	RootCmd.AddCommand(listCmd)
	RootCmd.AddCommand(rmCmd)
}

func Execute() {
	if err := RootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
