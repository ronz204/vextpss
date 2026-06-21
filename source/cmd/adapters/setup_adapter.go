package adapters

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"

	"vextpss/source/cmd/formatters"
	"vextpss/source/shared/storage"
)

func InitCmd(app App) *cobra.Command {
	return &cobra.Command{
		Use:   "init",
		Short: "Initialize the Vext environment",
		Long:  "Creates the config directory and database. Safe to run multiple times.",
		Args:  cobra.NoArgs,
		RunE: func(_ *cobra.Command, _ []string) error {
			db, err := storage.NewInitialiser(app.DBPath).Setup(context.Background())
			if err != nil {
				formatters.Error(err.Error())
				return err
			}
			_ = storage.Close(db)
			formatters.Success(fmt.Sprintf("Vext initialized at %s", app.DBPath))
			return nil
		},
	}
}
