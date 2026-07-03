package getsecrets

import (
	"github.com/spf13/cobra"

	"vextpss/source/funcs"
)

func NewCommand(deps funcs.Deps) *cobra.Command {
	return &cobra.Command{
		Use:   "get <name>",
		Short: "Get and display a secret's credentials",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return run(cmd.Context(), args[0], deps)
		},
	}
}
