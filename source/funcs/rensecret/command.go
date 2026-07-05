package rensecret

import (
	"github.com/spf13/cobra"

	"vextpss/source/funcs"
)

func NewCommand(deps funcs.Deps) *cobra.Command {
	return &cobra.Command{
		Use:   "rename <old> <new>",
		Short: "Rename an existing secret",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			return run(cmd.Context(), args[0], args[1], deps)
		},
	}
}
