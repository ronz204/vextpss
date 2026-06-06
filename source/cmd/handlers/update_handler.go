package handlers

import (
	"context"
	"fmt"

	"vextpss/source/cmd/helpers"
	"vextpss/source/core"
	"vextpss/source/pkg/apps"

	"github.com/spf13/cobra"
)

// UpdateHandler handles the `vext update <name>` command.
type UpdateHandler struct {
	uc       *apps.UpdateSecretUC
	prompter helpers.Prompter
}

func NewUpdateHandler(uc *apps.UpdateSecretUC, prompter helpers.Prompter) *UpdateHandler {
	return &UpdateHandler{uc: uc, prompter: prompter}
}

func (h *UpdateHandler) CobraCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "update <name>",
		Short: "Update an existing credential",
		Long:  "Replaces the password (and optionally the username) of a stored credential.",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return h.Handle(context.Background(), args[0])
		},
	}
}

func (h *UpdateHandler) Handle(ctx context.Context, name string) error {
	newUsername, err := h.prompter.ReadLine("New Username (leave blank to keep current): ")
	if err != nil {
		return fmt.Errorf("could not read username: %w", err)
	}

	newPassword, err := h.prompter.ReadPassword("New Password: ")
	if err != nil {
		return fmt.Errorf("could not read password: %w", err)
	}
	defer h.prompter.Zero(newPassword)

	masterPassword, err := h.prompter.ReadPassword("Master Password: ")
	if err != nil {
		return fmt.Errorf("could not read master password: %w", err)
	}
	defer h.prompter.Zero(masterPassword)

	req := apps.UpdateSecretRequest{
		Name:           name,
		MasterPassword: masterPassword,
		NewUsername:    newUsername,
		NewPassword:    newPassword,
	}

	if err := h.uc.Execute(ctx, req); err != nil {
		if core.IsNotFound(err) {
			fmt.Printf("[X] Error: no credential named %q found.\n", name)
			return nil
		}
		if core.IsDomainError(err) {
			fmt.Printf("[X] %s\n", err)
			return nil
		}
		fmt.Println("[X] An unexpected error occurred. Please try again.")
		return nil
	}

	fmt.Printf("[✓] Credential %q updated.\n", name)
	return nil
}
