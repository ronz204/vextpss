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

// ================================
// UpdCmd returns the cobra command for "vext upd <name>".
// The secret type is resolved from the existing record — no --type flag required.
// ================================
func UpdCmd(dbPath string, enc funcs.Encryptor, input funcs.Collector) *cobra.Command {
	return &cobra.Command{
		Use:   "update <name>",
		Short: "Update an existing secret",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runUpd(args[0], dbPath, enc, input)
		},
	}
}

func runUpd(name, dbPath string, enc funcs.Encryptor, input funcs.Collector) error {
	var secretType string
	if err := storage.WithRepo(dbPath, func(repo *storage.SecretRepository) error {
		existing, _, err := repo.GetByName(context.Background(), name)
		if err != nil {
			return err
		}
		secretType = existing.Type
		return nil
	}); err != nil {
		if errors.Is(err, sentinel.ErrSecretNotFound) {
			formatters.Error(fmt.Sprintf("no secret named %q found", name))
		} else {
			formatters.Error(err.Error())
		}
		return err
	}

	plaintext, err := input.Payload(secretType)
	defer memory.Cleaner(plaintext)
	if err != nil {
		formatters.Error(err.Error())
		return err
	}

	masterPassword, err := input.Master()
	defer memory.Cleaner(masterPassword)
	if err != nil {
		formatters.Error(err.Error())
		return err
	}

	if err := storage.WithRepo(dbPath, func(repo *storage.SecretRepository) error {
		return funcs.NewUpdateSecretFunc(repo, enc).Run(context.Background(), funcs.UpdateSecretDto{
			Name:           name,
			Type:           secretType,
			Plaintext:      plaintext,
			MasterPassword: masterPassword,
		})
	}); err != nil {
		formatters.Error(err.Error())
		return err
	}

	formatters.Success(fmt.Sprintf("Secret %q updated.", name))
	return nil
}
