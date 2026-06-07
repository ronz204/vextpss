package handlers

import (
	"context"
	"fmt"

	"vextpss/source/cmd/helpers"
	"vextpss/source/core"
	"vextpss/source/pkg/apps"

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
	var secretType string
	var generate bool
	var genLength int
	var genNoSymbols bool

	cmd := &cobra.Command{
		Use:   "add <name>",
		Short: "Store a new credential",
		Long:  "Interactively stores a new credential under the given name. Use --type to specify the secret type (account, credit).",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return h.Handle(context.Background(), args[0], secretType, generate, genLength, !genNoSymbols)
		},
	}
	cmd.Flags().StringVar(&secretType, "type", "account", "Secret type: account or credit")
	cmd.Flags().BoolVar(&generate, "generate", false, "Generate a random password instead of prompting (account only)")
	cmd.Flags().IntVar(&genLength, "gen-length", 20, "Length of the generated password")
	cmd.Flags().BoolVar(&genNoSymbols, "gen-no-symbols", false, "Exclude symbols from the generated password")
	return cmd
}

func (h *AddHandler) Handle(ctx context.Context, name string, secretType string, generate bool, genLength int, genSymbols bool) error {
	opts := helpers.CollectorOptions{
		Generate:   generate,
		GenLength:  genLength,
		GenSymbols: genSymbols,
	}

	collector, err := helpers.NewSecretCollector(secretType, opts)
	if err != nil {
		fmt.Printf("[X] %s\n", err)
		return nil
	}

	payload, err := collector.Collect(h.prompter)
	if err != nil {
		if core.IsDomainError(err) {
			fmt.Printf("[X] %s\n", err)
			return nil
		}
		return fmt.Errorf("input collection failed: %w", err)
	}

	masterPassword, err := h.prompter.ReadPassword("Master Password: ")
	if err != nil {
		return fmt.Errorf("could not read master password: %w", err)
	}
	defer h.prompter.Zero(masterPassword)

	req := apps.StoreSecretRequest{
		Name:           name,
		MasterPassword: masterPassword,
		Payload:        payload,
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
