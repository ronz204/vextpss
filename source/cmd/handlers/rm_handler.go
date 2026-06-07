package handlers

import (
	"context"
	"fmt"

	"vextpss/source/app"
	"vextpss/source/cmd/ui"

	"github.com/spf13/cobra"
)

// RmHandler handles the `vext rm <name>` command.
type RmHandler struct {
	uc       *app.DeleteSecretUC
	prompter ui.Prompter
}

func NewRmHandler(uc *app.DeleteSecretUC, prompter ui.Prompter) *RmHandler {
	return &RmHandler{uc: uc, prompter: prompter}
}

func (h *RmHandler) CobraCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "rm <name>",
		Short: "Delete a stored credential",
		Long:  "Permanently removes a stored credential by name after a confirmation prompt.",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return h.Handle(context.Background(), args[0])
		},
	}
}

func (h *RmHandler) Handle(ctx context.Context, name string) error {
	if !guardInit(h.uc != nil) {
		return nil
	}

	confirmed, err := h.prompter.Confirm(
		fmt.Sprintf("Are you sure you want to delete %q? This cannot be undone. [y/N]: ", name),
	)
	if err != nil {
		return fmt.Errorf("could not read confirmation: %w", err)
	}
	if !confirmed {
		fmt.Println("Aborted.")
		return nil
	}

	if printErr(h.uc.Execute(ctx, name), fmt.Sprintf("Error: no credential named %q found.", name)) {
		return nil
	}

	fmt.Printf("[✓] Credential %q deleted.\n", name)
	return nil
}
