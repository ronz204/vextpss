package cmd

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"

	"vextpss/source/cmd/formatters"
	"vextpss/source/shared/storage"
)

func InitCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "init",
		Short: "Initialize the Vext environment",
		Long:  "Creates the config directory and database. Safe to run multiple times.",
		Args:  cobra.NoArgs,
		RunE: func(_ *cobra.Command, _ []string) error {
			dbPath, err := storage.DBPath()
			if err != nil {
				formatters.Error(err.Error())
				return err
			}
			if err := storage.SetupDB(context.Background(), dbPath); err != nil {
				formatters.Error(err.Error())
				return err
			}
			formatters.Success(fmt.Sprintf("Vext initialized at %s", dbPath))
			return nil
		},
	}
}
