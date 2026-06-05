package handlers

import (
	"context"
	"fmt"

	"vextpss/source/cmd/helpers"
	"vextpss/source/pkg/apps"

	"github.com/spf13/cobra"
)

// ListHandler handles the `vext list` command.
type ListHandler struct {
	uc *apps.ListSecretsUC
}

func NewListHandler(uc *apps.ListSecretsUC) *ListHandler {
	return &ListHandler{uc: uc}
}

func (h *ListHandler) CobraCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List all stored secrets",
		Long:  "Displays a table of all stored secret names and types. Does not require the master password.",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			return h.Handle(context.Background())
		},
	}
}

func (h *ListHandler) Handle(ctx context.Context) error {
	secrets, err := h.uc.Execute(ctx)
	if err != nil {
		fmt.Printf("[X] Could not list secrets: %s\n", err)
		return nil
	}

	helpers.PrintSecretList(secrets)
	return nil
}
