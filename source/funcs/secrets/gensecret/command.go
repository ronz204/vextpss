package gensecret

import (
	"github.com/spf13/cobra"
)

func NewCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "gen",
		Short: "Generate a cryptographically secure password",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			length, _ := cmd.Flags().GetInt("length")
			symbols, _ := cmd.Flags().GetBool("symbols")
			return run(length, symbols)
		},
	}
	cmd.Flags().IntP("length", "l", 20, "Password length")
	cmd.Flags().BoolP("symbols", "s", false, "Include symbols")
	return cmd
}
