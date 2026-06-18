package cmd

import (
	"github.com/spf13/cobra"

	"vextpss/source/cmd/adapters"
	"vextpss/source/cmd/collectors"
	"vextpss/source/shared"
	"vextpss/source/shared/crypto"
	"vextpss/source/shared/storage"
)

// Execute wires all dependencies and runs the root command.
func Execute() error {
	path := shared.DBPath()
	enc := crypto.NewAESGCMEncryptor(crypto.DefaultConfig())
	prompter := collectors.NewTerminalPrompter()
	collector := collectors.NewTerminalCollector(prompter)

	root := &cobra.Command{
		Use:           shared.AppName,
		Short:         "A local-first CLI password manager",
		SilenceErrors: true,
		SilenceUsage:  true,
	}

	root.AddCommand(
		adapters.InitCmd(storage.NewInitialiser(path)),
		adapters.AddCmd(path, enc, collector),
		adapters.UpdCmd(path, enc, collector),
		adapters.RmCmd(path, collector),
	)

	return root.Execute()
}
