package usespace

import (
	"github.com/spf13/cobra"

	"vextpss/source/funcs"
)

func NewCommand(deps funcs.Deps) *cobra.Command {
	return &cobra.Command{
		Use:   "use [space]",
		Short: "Switch to a different space (no args: show active)",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) == 0 {
				return showActive(deps)
			}
			return run(cmd.Context(), args[0], deps)
		},
	}
}
