package adapters

import (
	"context"
	"errors"
	"fmt"

	"github.com/spf13/cobra"

	"vextpss/source/cmd/formatters"
	"vextpss/source/core"
	"vextpss/source/funcs"
	"vextpss/source/shared"
	"vextpss/source/shared/passgen"
	"vextpss/source/shared/sentinel"
	"vextpss/source/shared/storage"
)

// ================================
// ImportCmd returns the cobra command for "vext import <file>".
// ================================
func ImportCmd(deps shared.AppDeps) *cobra.Command {
	return &cobra.Command{
		Use:   "import <file>",
		Short: "Import secrets from an encrypted export file",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runImport(args[0], deps)
		},
	}
}

func runImport(filePath string, deps shared.AppDeps) error {
	masterPassword, err := deps.Collector.CollectMaster()
	defer passgen.Cleaner(masterPassword)
	if err != nil {
		formatters.Error(err.Error())
		return err
	}

	var result funcs.ImportResult
	err = storage.WithRepo(deps.DBPath, func(repo core.SecretRepository) error {
		var runErr error
		result, runErr = funcs.NewImportSecretsFunc(repo, deps.Enc).Run(context.Background(), funcs.ImportSecretsDto{
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
