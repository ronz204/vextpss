package addsecret

import (
	"github.com/spf13/cobra"

	"vextpss/source/funcs"
	"vextpss/source/secrets"
)

func NewCommand(deps funcs.Deps) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "add <name>",
		Short: "Add a new secret",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			secretType, _ := cmd.Flags().GetString("type")
			return run(cmd.Context(), args[0], secretType, deps)
		},
	}
	cmd.Flags().StringP("type", "t", secrets.TypeAccount, "Secret type (account, finance)")
	return cmd
}
