package cmd

import (
	"github.com/spf13/cobra"

	"vextpss/source/cmd/adapters"
	"vextpss/source/shared/storage"
)

// Execute builds AppDeps, registers all commands, and runs the root cobra command.
func Execute() error {
	deps := Build()
	dbPath := storage.DBPath()

	root := &cobra.Command{
		Use:           "vext",
		Short:         "A local-first CLI password manager",
		SilenceErrors: true,
		SilenceUsage:  true,
	}

	root.AddCommand(
		adapters.InitCmd(storage.NewInitialiser(dbPath)),
		adapters.AddCmd(dbPath, deps.Encryptor, deps.Collector),
		adapters.GetCmd(dbPath, deps.Encryptor, deps.Collector),
		adapters.ListCmd(dbPath),
		adapters.UpdCmd(dbPath, deps.Encryptor, deps.Collector),
		adapters.RmCmd(dbPath, deps.Collector),
		adapters.ExportCmd(dbPath, deps.Encryptor, deps.Collector),
		adapters.ImportCmd(dbPath, deps.Encryptor, deps.Collector),
		adapters.GenCmd(),
	)

	return root.Execute()
}
