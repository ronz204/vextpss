package cmd

import (
	"github.com/spf13/cobra"

	"vextpss/source/cmd/adapters"
	"vextpss/source/cmd/formatters"
)

func Execute() error {
	var (
		app     adapters.App
		cleanup = func() {}
	)
	defer func() { cleanup() }()

	root := &cobra.Command{
		Use:           "vext",
		Short:         "Encrypted secrets manager",
		SilenceUsage:  true,
		SilenceErrors: true,
		PersistentPreRunE: func(cmd *cobra.Command, _ []string) error {
			if cmd.Name() == "init" {
				return nil
			}

			built, cl, err := adapters.Build()
			if err != nil {
				formatters.Error("could not open database — run `vext init` first")
				return err
			}

			app = built
			cleanup = cl
			return nil
		},
	}

	root.AddCommand(
		adapters.InitCmd(),
		adapters.AddCmd(&app),
		adapters.GetCmd(&app),
		adapters.ListCmd(&app),
		adapters.RmCmd(&app),
	)

	return root.Execute()
}
