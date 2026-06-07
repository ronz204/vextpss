package handlers

import (
	"context"
	"fmt"

	"vextpss/source/app"
	"vextpss/source/cmd/ui"

	"github.com/spf13/cobra"
)

// ImportHandler handles the `vext import <file>` command.
type ImportHandler struct {
	uc       *app.ImportSecretsUC
	prompter ui.Prompter
}

func NewImportHandler(uc *app.ImportSecretsUC, prompter ui.Prompter) *ImportHandler {
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
	if !guardInit(h.uc != nil) {
		return nil
	}

	masterPassword, err := h.prompter.ReadPassword("Master Password: ")
	if err != nil {
		return fmt.Errorf("could not read master password: %w", err)
	}
	defer h.prompter.Zero(masterPassword)

	req := app.ImportSecretsRequest{
		MasterPassword: masterPassword,
		InputPath:      inputPath,
	}

	resp, err := h.uc.Execute(ctx, req)
	if printErr(err, "") {
		return nil
	}

	msg := fmt.Sprintf("[✓] %d secret(s) imported", resp.Imported)
	if resp.Skipped > 0 {
		msg += fmt.Sprintf(", %d skipped (name already exists)", resp.Skipped)
	}
	fmt.Println(msg + ".")
	return nil
}
