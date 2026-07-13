package listspaces

import (
	"github.com/spf13/cobra"

	"vextpss/source/funcs"
)

func NewCommand(deps funcs.Deps) *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List all spaces",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			return Run(cmd.Context(), deps)
		},
	}
}
