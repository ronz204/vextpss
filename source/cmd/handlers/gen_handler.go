package handlers

import (
	"fmt"

	"vextpss/source/pkg/shared"

	"github.com/spf13/cobra"
)

// GenHandler handles the `vext gen` command.
type GenHandler struct{}

func NewGenHandler() *GenHandler { return &GenHandler{} }

func (h *GenHandler) CobraCommand() *cobra.Command {
	var length int
	var noSymbols bool

	cmd := &cobra.Command{
		Use:   "gen",
		Short: "Generate a secure random password",
		Long:  "Generates a cryptographically secure random password using crypto/rand.",
		RunE: func(cmd *cobra.Command, args []string) error {
			return h.Handle(length, !noSymbols)
		},
	}
	cmd.Flags().IntVarP(&length, "length", "l", 20, "Password length")
	cmd.Flags().BoolVar(&noSymbols, "no-symbols", false, "Exclude symbols from the character set")
	return cmd
}

func (h *GenHandler) Handle(length int, useSymbols bool) error {
	if length < 1 {
		fmt.Println("[X] Error: length must be at least 1.")
		return nil
	}
	pw, err := shared.GeneratePassword(length, useSymbols)
	if err != nil {
		return fmt.Errorf("failed to generate password: %w", err)
	}
	fmt.Println(pw)
	return nil
}
