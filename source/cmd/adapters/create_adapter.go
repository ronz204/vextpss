package adapters

import (
	"context"
	"errors"
	"fmt"

	"github.com/spf13/cobra"

	"vextpss/source/cmd/formatters"
	"vextpss/source/funcs"
	"vextpss/source/secrets"
	"vextpss/source/shared/memory"
	"vextpss/source/shared/sentinel"
	"vextpss/source/shared/storage"
)

// ================================
// AddCmd returns the cobra command for "vext add <name>".
// ================================
func AddCmd(dbPath string, enc funcs.Encryptor, input funcs.Collector) *cobra.Command {
	var secretType string

	cmd := &cobra.Command{
		Use:   "add <name>",
		Short: "Store a new secret",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runAdd(args[0], secretType, dbPath, enc, input)
		},
	}

	cmd.Flags().StringVarP(&secretType, "type", "t", secrets.TypeAccount, `secret type: "account" or "finance"`)
	return cmd
}

func runAdd(name, secretType, dbPath string, enc funcs.Encryptor, input funcs.Collector) error {
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

	err = storage.WithRepo(dbPath, func(repo *storage.SecretRepository) error {
		return funcs.NewCreateSecretFunc(repo, enc).Run(context.Background(), funcs.CreateSecretDto{
			Name:           name,
			Type:           secretType,
			Plaintext:      plaintext,
			MasterPassword: masterPassword,
		})
	})

	if err != nil {
		if errors.Is(err, sentinel.ErrAlreadyExists) {
			formatters.Error(fmt.Sprintf("a credential named %q already exists. Use `vext upd` to modify it.", name))
		} else {
			formatters.Error(err.Error())
		}
		return err
	}

	formatters.Success(fmt.Sprintf("Credential %q saved.", name))
	return nil
}
