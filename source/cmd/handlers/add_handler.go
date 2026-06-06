package handlers

import (
	"context"
	"fmt"

	"vextpss/source/cmd/helpers"
	"vextpss/source/core"
	"vextpss/source/pkg/apps"
	"vextpss/source/pkg/shared"

	"github.com/spf13/cobra"
)

// AddHandler handles the `vext add <name>` command.
type AddHandler struct {
	uc       *apps.StoreSecretUC
	prompter helpers.Prompter
}

func NewAddHandler(uc *apps.StoreSecretUC, prompter helpers.Prompter) *AddHandler {
	return &AddHandler{uc: uc, prompter: prompter}
}

func (h *AddHandler) CobraCommand() *cobra.Command {
	var generate bool
	var genLength int
	var genNoSymbols bool

	cmd := &cobra.Command{
		Use:   "add <name>",
		Short: "Store a new credential",
		Long:  "Interactively stores a new account credential under the given name.",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return h.Handle(context.Background(), args[0], generate, genLength, !genNoSymbols)
		},
	}
	cmd.Flags().BoolVar(&generate, "generate", false, "Generate a random password instead of prompting")
	cmd.Flags().IntVar(&genLength, "gen-length", 20, "Length of the generated password")
	cmd.Flags().BoolVar(&genNoSymbols, "gen-no-symbols", false, "Exclude symbols from the generated password")
	return cmd
}

func (h *AddHandler) Handle(ctx context.Context, name string, generate bool, genLength int, genSymbols bool) error {
	username, err := h.prompter.ReadLine("Username: ")
	if err != nil {
		return fmt.Errorf("could not read username: %w", err)
	}

	var password []byte
	if generate {
		pw, err := shared.GeneratePassword(genLength, genSymbols)
		if err != nil {
			return fmt.Errorf("failed to generate password: %w", err)
		}
		password = []byte(pw)
		fmt.Printf("Generated password: %s\n", pw)
	} else {
		password, err = h.prompter.ReadPassword("Password: ")
		if err != nil {
			return fmt.Errorf("could not read password: %w", err)
		}
	}
	defer h.prompter.Zero(password)

	masterPassword, err := h.prompter.ReadPassword("Master Password: ")
	if err != nil {
		return fmt.Errorf("could not read master password: %w", err)
	}
	defer h.prompter.Zero(masterPassword)

	req := apps.StoreSecretRequest{
		Name:            name,
		Type:            "account",
		MasterPassword:  masterPassword,
		AccountUsername: username,
		AccountPassword: password,
	}

	if err := h.uc.Execute(ctx, req); err != nil {
		if core.IsAlreadyExists(err) {
			fmt.Printf("[X] Error: a credential named %q already exists. Use `vext update` to modify it.\n", name)
			return nil
		}
		if core.IsDomainError(err) {
			fmt.Printf("[X] %s\n", err)
			return nil
		}
		fmt.Println("[X] An unexpected error occurred. Please try again.")
		return nil
	}

	fmt.Printf("[✓] Credential %q saved.\n", name)
	return nil
}
