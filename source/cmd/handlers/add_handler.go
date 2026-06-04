package handlers

import (
	"context"
	"fmt"
	"vextpss/source/cmd/ui"
	"vextpss/source/pkg/application"
	"vextpss/source/pkg/domain"

	"github.com/spf13/cobra"
)

// AddHandler handles the `vext add <name>` command.
type AddHandler struct {
	uc       *application.StoreSecretUC
	prompter ui.Prompter
}

func NewAddHandler(uc *application.StoreSecretUC, prompter ui.Prompter) *AddHandler {
	return &AddHandler{uc: uc, prompter: prompter}
}

func (h *AddHandler) CobraCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "add <name>",
		Short: "Store a new credential",
		Long:  "Interactively stores a new account credential under the given name.",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return h.Handle(context.Background(), args[0])
		},
	}
}

func (h *AddHandler) Handle(ctx context.Context, name string) error {
	username, err := h.prompter.ReadLine("Username: ")
	if err != nil {
		return fmt.Errorf("could not read username: %w", err)
	}

	password, err := h.prompter.ReadPassword("Password: ")
	if err != nil {
		return fmt.Errorf("could not read password: %w", err)
	}
	defer h.prompter.Zero(password)

	masterPassword, err := h.prompter.ReadPassword("Master Password: ")
	if err != nil {
		return fmt.Errorf("could not read master password: %w", err)
	}
	defer h.prompter.Zero(masterPassword)

	req := application.StoreSecretRequest{
		Name:            name,
		Type:            "account",
		MasterPassword:  masterPassword,
		AccountUsername: username,
		AccountPassword: password,
	}

	if err := h.uc.Execute(ctx, req); err != nil {
		if domain.IsAlreadyExists(err) {
			fmt.Printf("[X] Error: a credential named %q already exists. Use `vext rm` then `vext add` to replace it.\n", name)
			return nil
		}
		if domain.IsDomainError(err) {
			fmt.Printf("[X] %s\n", err)
			return nil
		}
		fmt.Println("[X] An unexpected error occurred. Please try again.")
		return nil
	}

	fmt.Printf("[✓] Credential %q saved.\n", name)
	return nil
}
