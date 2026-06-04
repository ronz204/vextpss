package handlers

import (
	"context"
	"fmt"
	"vextpss/source/pkg/application"

	"github.com/spf13/cobra"
)

// InitHandler handles the `vext init` command.
type InitHandler struct {
	uc *application.InitStorageUC
}

func NewInitHandler(uc *application.InitStorageUC) *InitHandler {
	return &InitHandler{uc: uc}
}

func (h *InitHandler) CobraCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "init",
		Short: "Initialize the Vext environment",
		Long:  "Creates the config directory and database on first use. Safe to run multiple times.",
		RunE: func(cmd *cobra.Command, args []string) error {
			return h.Handle(context.Background())
		},
	}
}

func (h *InitHandler) Handle(ctx context.Context) error {
	if err := h.uc.Execute(ctx); err != nil {
		fmt.Printf("[X] %s\n", err)
		return nil
	}
	fmt.Println("[✓] Vext initialized.")
	return nil
}
