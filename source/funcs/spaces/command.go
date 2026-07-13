package spaces

import (
	"github.com/spf13/cobra"

	"vextpss/source/funcs"
	"vextpss/source/funcs/spaces/addspace"
	"vextpss/source/funcs/spaces/dropspace"
	"vextpss/source/funcs/spaces/listspaces"
	"vextpss/source/funcs/spaces/renspace"
)

func NewCommand(deps funcs.Deps) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "spaces",
		Short: "Manage spaces",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			return listspaces.Run(cmd.Context(), deps)
		},
	}
	cmd.AddCommand(addspace.NewCommand(deps))
	cmd.AddCommand(dropspace.NewCommand(deps))
	cmd.AddCommand(renspace.NewCommand(deps))
	cmd.AddCommand(listspaces.NewCommand(deps))
	return cmd
}
