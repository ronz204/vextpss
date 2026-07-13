package dropspace

import (
	"github.com/spf13/cobra"

	"vextpss/source/funcs"
)

func NewCommand(deps funcs.Deps) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "drop <name>",
		Short: "Delete a space",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			force, _ := cmd.Flags().GetBool("force")
			return run(cmd.Context(), args[0], force, deps)
		},
	}
	cmd.Flags().BoolP("force", "f", false, "Also delete all secrets inside the space")
	return cmd
}
