package handlers

import (
	"context"
	"fmt"

	"vextpss/source/app"
	"vextpss/source/cmd/ui"

	"github.com/spf13/cobra"
)

// GetHandler handles the `vext get <name>` command.
type GetHandler struct {
	uc       *app.RetrieveSecretUC
	prompter ui.Prompter
}

func NewGetHandler(uc *app.RetrieveSecretUC, prompter ui.Prompter) *GetHandler {
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
	if !guardInit(h.uc != nil) {
		return nil
	}

	masterPassword, err := h.prompter.ReadPassword("Master Password: ")
	if err != nil {
		return fmt.Errorf("could not read master password: %w", err)
	}
	defer h.prompter.Zero(masterPassword)

	req := app.RetrieveSecretRequest{
		Name:           name,
		MasterPassword: masterPassword,
	}

	resp, err := h.uc.Execute(ctx, req)
	if printErr(err, fmt.Sprintf("Error: no credential named %q found.", name)) {
		return nil
	}

	ui.PrintSecret(resp.Name, resp.Payload)
	return nil
}
