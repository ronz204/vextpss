package handlers

import (
	"context"
	"fmt"

	"vextpss/source/cmd/helpers"
	"vextpss/source/core"
	"vextpss/source/pkg/apps"

	"github.com/spf13/cobra"
)

// ImportHandler handles the `vext import <file>` command.
type ImportHandler struct {
	uc       *apps.ImportSecretsUC
	prompter helpers.Prompter
}

func NewImportHandler(uc *apps.ImportSecretsUC, prompter helpers.Prompter) *ImportHandler {
	return &ImportHandler{uc: uc, prompter: prompter}
}

func (h *ImportHandler) CobraCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "import <file>",
		Short: "Import secrets from an encrypted backup file",
		Long:  "Restores secrets from a .vext file produced by `vext export`.",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return h.Handle(context.Background(), args[0])
		},
	}
}

func (h *ImportHandler) Handle(ctx context.Context, inputPath string) error {
	masterPassword, err := h.prompter.ReadPassword("Master Password: ")
	if err != nil {
		return fmt.Errorf("could not read master password: %w", err)
	}
	defer h.prompter.Zero(masterPassword)

	req := apps.ImportSecretsRequest{
		MasterPassword: masterPassword,
		InputPath:      inputPath,
	}

	resp, err := h.uc.Execute(ctx, req)
	if err != nil {
		if core.IsDomainError(err) {
			fmt.Printf("[X] %s\n", err)
			return nil
		}
		fmt.Println("[X] An unexpected error occurred. Please try again.")
		return nil
	}

	msg := fmt.Sprintf("[✓] %d secret(s) imported", resp.Imported)
	if resp.Skipped > 0 {
		msg += fmt.Sprintf(", %d skipped (name already exists)", resp.Skipped)
	}
	fmt.Println(msg + ".")
	return nil
}
