package adapters

import (
	"context"
	"errors"
	"fmt"

	"github.com/spf13/cobra"

	"vextpss/source/cmd/formatters"
	"vextpss/source/funcs"
	"vextpss/source/shared/memory"
	"vextpss/source/shared/sentinel"
	"vextpss/source/shared/storage"
)

func ImportCmd(d Deps) *cobra.Command {
	return &cobra.Command{
		Use:   "import <file>",
		Short: "Import secrets from an encrypted export file",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runImport(args[0], d)
		},
	}
}

func runImport(filePath string, d Deps) error {
	masterPassword, err := d.Collector.Master()
	defer memory.Cleaner(masterPassword)
	if err != nil {
		formatters.Error(err.Error())
		return err
	}

	var result funcs.ImportResult
	err = storage.WithRepo(d.DBPath, func(repo *storage.SecretRepository) error {
		var runErr error
		result, runErr = funcs.NewImportSecretsFunc(repo, d.Encryptor).Run(context.Background(), funcs.ImportSecretsDto{
			FilePath:       filePath,
			MasterPassword: masterPassword,
		})
		return runErr
	})
	if err != nil {
		if errors.Is(err, sentinel.ErrDecryptionFailed) {
			formatters.Error("wrong master password or corrupted export file")
		} else {
			formatters.Error(err.Error())
		}
		return err
	}

	formatters.Success(fmt.Sprintf("Imported %d secret(s), skipped %d duplicate(s).", result.Imported, result.Skipped))
	return nil
}
