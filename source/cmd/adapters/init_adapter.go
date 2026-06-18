package adapters

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"

	"vextpss/source/cmd/formatters"
	"vextpss/source/shared/storage"
)

// ================================
// InitCmd returns the cobra command for "vext init".
// ================================
func InitCmd(i *storage.Initialiser) *cobra.Command {
	return &cobra.Command{
		Use:   "init",
		Short: "Initialize Vext on this machine",
		Long:  "Creates the config directory and database. Safe to run multiple times.",
		RunE: func(cmd *cobra.Command, args []string) error {
			db, err := i.Setup(context.Background())
			if err != nil {
				formatters.Error(err.Error())
				return err
			}
			storage.Close(db)
			formatters.Success(fmt.Sprintf("Vext initialized at %s", i.DBPath()))
			return nil
		},
	}
}
