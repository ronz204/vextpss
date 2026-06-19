package adapters

import (
	"context"

	"github.com/spf13/cobra"

	"vextpss/source/cmd/formatters"
	"vextpss/source/funcs"
	"vextpss/source/shared/storage"
)

// ================================
// ListCmd returns the cobra command for "vext list".
// No master password required — payloads are never touched.
// ================================
func ListCmd(dbPath string) *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List all stored secrets",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runList(dbPath)
		},
	}
}

func runList(dbPath string) error {
	err := storage.WithRepo(dbPath, func(repo *storage.SecretRepository) error {
		all, err := funcs.NewRetrieveSecretsFunc(repo).Run(context.Background())
		if err != nil {
			return err
		}
		if len(all) == 0 {
			formatters.Info("No secrets stored yet. Use 'vext add' to get started.")
			return nil
		}
		formatters.PrintTabTable(all)
		return nil
	})
	if err != nil {
		formatters.Error(err.Error())
		return err
	}
	return nil
}
