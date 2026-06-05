package handlers

import (
	"context"
	"fmt"

	"vextpss/source/cmd/helpers"
	"vextpss/source/core"
	"vextpss/source/pkg/apps"

	"github.com/spf13/cobra"
)

// GetHandler handles the `vext get <name>` command.
type GetHandler struct {
	uc       *apps.RetrieveSecretUC
	prompter helpers.Prompter
}

func NewGetHandler(uc *apps.RetrieveSecretUC, prompter helpers.Prompter) *GetHandler {
	return &GetHandler{uc: uc, prompter: prompter}
}

func (h *GetHandler) CobraCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "get <name>",
		Short: "Retrieve and display a stored credential",
		Long:  "Looks up a stored credential by name, decrypts it, and displays its fields.",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return h.Handle(context.Background(), args[0])
		},
	}
}

func (h *GetHandler) Handle(ctx context.Context, name string) error {
	masterPassword, err := h.prompter.ReadPassword("Master Password: ")
	if err != nil {
		return fmt.Errorf("could not read master password: %w", err)
	}
	defer h.prompter.Zero(masterPassword)

	req := apps.RetrieveSecretRequest{
		Name:           name,
		MasterPassword: masterPassword,
	}

	resp, err := h.uc.Execute(ctx, req)
	if err != nil {
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

	helpers.PrintSecret(resp.Name, resp.Payload)
	return nil
}
