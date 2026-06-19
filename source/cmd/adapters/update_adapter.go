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

func UpdCmd(d Deps) *cobra.Command {
	return &cobra.Command{
		Use:   "update <name>",
		Short: "Update an existing secret",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runUpd(args[0], d)
		},
	}
}

func runUpd(name string, d Deps) error {
	var secretType string
	if err := storage.WithRepo(d.DBPath, func(repo *storage.SecretRepository) error {
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

	plaintext, err := d.Collector.Payload(secretType)
	defer memory.Cleaner(plaintext)
	if err != nil {
		formatters.Error(err.Error())
		return err
	}

	masterPassword, err := d.Collector.Master()
	defer memory.Cleaner(masterPassword)
	if err != nil {
		formatters.Error(err.Error())
		return err
	}

	if err := storage.WithRepo(d.DBPath, func(repo *storage.SecretRepository) error {
		return funcs.NewUpdateSecretFunc(repo, d.Encryptor).Run(context.Background(), funcs.UpdateSecretDto{
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
