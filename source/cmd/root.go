package cmd

import (
	"github.com/spf13/cobra"

	"vextpss/source/cmd/adapters"
	"vextpss/source/cmd/formatters"
)

func Execute() error {
	app, cleanup, err := adapters.Build()
	if err != nil {
		formatters.Error(err.Error())
		return err
	}
	defer cleanup()

	root := &cobra.Command{
		Use:           "vext",
		Short:         "Encrypted secrets manager",
		SilenceUsage:  true,
		SilenceErrors: true,
	}

	root.AddCommand(
		adapters.InitCmd(app),
		adapters.AddCmd(app),
		adapters.GetCmd(app),
		adapters.ListCmd(app),
		adapters.RmCmd(app),
	)

	return root.Execute()
}
