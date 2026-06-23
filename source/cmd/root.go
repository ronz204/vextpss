package cmd

import "github.com/spf13/cobra"

func Execute() error {
	root := &cobra.Command{
		Use:           "vext",
		Short:         "Encrypted secrets manager",
		SilenceUsage:  true,
		SilenceErrors: true,
	}
	root.AddCommand(
		InitCmd(),
		AddCmd(),
		GetCmd(),
		ListCmd(),
		RmCmd(),
	)
	return root.Execute()
}
