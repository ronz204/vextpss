package cmd

import (
	"github.com/spf13/cobra"

	"vextpss/source/cmd/adapters"
	"vextpss/source/cmd/collectors"
	"vextpss/source/shared"
	"vextpss/source/shared/crypto"
	"vextpss/source/shared/storage"
)

// Execute builds AppDeps, registers all commands, and runs the root cobra command.
func Execute() error {
	deps := shared.AppDeps{
		DBPath:    shared.DBPath(),
		Enc:       crypto.NewAESGCMEncryptor(crypto.DefaultConfig()),
		Collector: collectors.NewTerminalCollector(collectors.NewTerminalPrompter()),
	}

	root := &cobra.Command{
		Use:           shared.AppName,
		Short:         "A local-first CLI password manager",
		SilenceErrors: true,
		SilenceUsage:  true,
	}

	root.AddCommand(
		adapters.InitCmd(storage.NewInitialiser(deps.DBPath)),
		adapters.AddCmd(deps),
		adapters.UpdCmd(deps),
		adapters.RmCmd(deps),
		adapters.ExportCmd(deps),
		adapters.ImportCmd(deps),
	)

	return root.Execute()
}
