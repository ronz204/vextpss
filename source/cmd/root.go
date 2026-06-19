package cmd

import (
	"github.com/spf13/cobra"

	"vextpss/source/cmd/adapters"
	"vextpss/source/shared/storage"
)

func Execute() error {
	deps := adapters.BuildDeps()

	root := &cobra.Command{
		Use:           "vext",
		Short:         "A local-first CLI password manager",
		SilenceErrors: true,
		SilenceUsage:  true,
	}

	root.AddCommand(
		adapters.InitCmd(storage.NewInitialiser(deps.DBPath)),
		adapters.AddCmd(deps),
		adapters.GetCmd(deps),
		adapters.ListCmd(deps),
		adapters.UpdCmd(deps),
		adapters.RmCmd(deps),
		adapters.ExportCmd(deps),
		adapters.ImportCmd(deps),
		adapters.GenCmd(),
	)

	return root.Execute()
}
