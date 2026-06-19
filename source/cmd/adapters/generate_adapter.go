package adapters

import (
	"fmt"

	"github.com/spf13/cobra"

	"vextpss/source/cmd/formatters"
	"vextpss/source/shared/passgen"
)

// ================================
// GenCmd returns the cobra command for "vext gen".
// ================================
func GenCmd() *cobra.Command {
	var length int
	var noSymbols bool

	cmd := &cobra.Command{
		Use:   "gen",
		Short: "Generate a cryptographically secure password",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runGen(length, noSymbols)
		},
	}

	cmd.Flags().IntVarP(&length, "length", "l", 20, "Password length")
	cmd.Flags().BoolVar(&noSymbols, "no-symbols", false, "Exclude symbols")
	return cmd
}

func runGen(length int, noSymbols bool) error {
	if length < 1 {
		formatters.Error("length must be at least 1")
		return fmt.Errorf("invalid length: %d", length)
	}

	pwd, err := passgen.Generate(length, !noSymbols)
	if err != nil {
		formatters.Error(err.Error())
		return err
	}

	fmt.Println(pwd)
	return nil
}
