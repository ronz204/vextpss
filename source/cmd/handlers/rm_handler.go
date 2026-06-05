package handlers

import (
	"context"
	"fmt"

	"vextpss/source/cmd/helpers"
	"vextpss/source/core"
	"vextpss/source/pkg/apps"

	"github.com/spf13/cobra"
)

// RmHandler handles the `vext rm <name>` command.
type RmHandler struct {
	uc       *apps.DeleteSecretUC
	prompter helpers.Prompter
}

func NewRmHandler(uc *apps.DeleteSecretUC, prompter helpers.Prompter) *RmHandler {
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

	if err := h.uc.Execute(ctx, name); err != nil {
		if core.IsNotFound(err) {
			fmt.Printf("[X] Error: no credential named %q found.\n", name)
			return nil
		}
		fmt.Printf("[X] Could not delete credential: %s\n", err)
		return nil
	}

	fmt.Printf("[✓] Credential %q deleted.\n", name)
	return nil
}
