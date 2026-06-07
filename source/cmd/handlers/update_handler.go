package handlers

import (
	"context"
	"fmt"

	"vextpss/source/app"
	"vextpss/source/cmd/ui"
	"vextpss/source/core"

	"github.com/spf13/cobra"
)

// UpdateHandler handles the `vext update <name>` command.
type UpdateHandler struct {
	retrieveUC *app.RetrieveSecretUC
	updateUC   *app.UpdateSecretUC
	prompter   ui.Prompter
}

func NewUpdateHandler(retrieveUC *app.RetrieveSecretUC, updateUC *app.UpdateSecretUC, prompter ui.Prompter) *UpdateHandler {
	return &UpdateHandler{retrieveUC: retrieveUC, updateUC: updateUC, prompter: prompter}
}

func (h *UpdateHandler) CobraCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "update <name>",
		Short: "Update an existing credential",
		Long:  "Replaces the stored fields of a credential. Prompts for the master password to verify access, then collects new values.",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return h.Handle(context.Background(), args[0])
		},
	}
}

func (h *UpdateHandler) Handle(ctx context.Context, name string) error {
	if !guardInit(h.updateUC != nil) {
		return nil
	}

	masterPassword, err := h.prompter.ReadPassword("Master Password: ")
	if err != nil {
		return fmt.Errorf("could not read master password: %w", err)
	}
	defer h.prompter.Zero(masterPassword)

	// Verify master password and retrieve current type before prompting for new values.
	retrieveResp, err := h.retrieveUC.Execute(ctx, app.RetrieveSecretRequest{
		Name:           name,
		MasterPassword: masterPassword,
	})
	if printErr(err, fmt.Sprintf("Error: no credential named %q found.", name)) {
		return nil
	}

	// Use the type-specific collector to gather new field values.
	collector, err := ui.NewSecretCollector(retrieveResp.Type, ui.CollectorOptions{})
	if err != nil {
		fmt.Printf("[X] %s\n", err)
		return nil
	}

	newPayload, err := collector.Collect(h.prompter)
	if err != nil {
		if core.IsDomainError(err) {
			fmt.Printf("[X] %s\n", err)
			return nil
		}
		return fmt.Errorf("input collection failed: %w", err)
	}

	// Each payload type decides what merging means (e.g. AccountSecret keeps the current
	// username when the user leaves the field blank; CreditSecret replaces all fields).
	newPayload.MergeFrom(retrieveResp.Payload)

	err = h.updateUC.Execute(ctx, app.UpdateSecretRequest{
		Name:           name,
		MasterPassword: masterPassword,
		NewPayload:     newPayload,
	})
	if printErr(err, fmt.Sprintf("Error: no credential named %q found.", name)) {
		return nil
	}

	fmt.Printf("[✓] Credential %q updated.\n", name)
	return nil
}
