package handlers

import (
	"context"
	"fmt"
	"time"

	"vextpss/source/cmd/helpers"
	"vextpss/source/core"
	"vextpss/source/pkg/apps"

	"github.com/spf13/cobra"
)

// ExportHandler handles the `vext export` command.
type ExportHandler struct {
	uc       *apps.ExportSecretsUC
	prompter helpers.Prompter
}

func NewExportHandler(uc *apps.ExportSecretsUC, prompter helpers.Prompter) *ExportHandler {
	return &ExportHandler{uc: uc, prompter: prompter}
}

func (h *ExportHandler) CobraCommand() *cobra.Command {
	var outputPath string

	cmd := &cobra.Command{
		Use:   "export",
		Short: "Export all secrets to an encrypted backup file",
		Long:  "Decrypts all stored secrets and re-encrypts them as a single portable .vext file.",
		RunE: func(cmd *cobra.Command, args []string) error {
			if outputPath == "" {
				outputPath = fmt.Sprintf("vext-backup-%s.vext", time.Now().Format("20060102-150405"))
			}
			return h.Handle(context.Background(), outputPath)
		},
	}
	cmd.Flags().StringVarP(&outputPath, "out", "o", "", "Output file path (default: vext-backup-<timestamp>.vext)")
	return cmd
}

func (h *ExportHandler) Handle(ctx context.Context, outputPath string) error {
	masterPassword, err := h.prompter.ReadPassword("Master Password: ")
	if err != nil {
		return fmt.Errorf("could not read master password: %w", err)
	}
	defer h.prompter.Zero(masterPassword)

	req := apps.ExportSecretsRequest{
		MasterPassword: masterPassword,
		OutputPath:     outputPath,
	}

	count, err := h.uc.Execute(ctx, req)
	if err != nil {
		if core.IsDomainError(err) {
			fmt.Printf("[X] %s\n", err)
			return nil
		}
		fmt.Println("[X] An unexpected error occurred. Please try again.")
		return nil
	}

	fmt.Printf("[✓] %d secret(s) exported to %q.\n", count, outputPath)
	return nil
}
