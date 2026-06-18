package adapters

import (
	"context"
	"errors"
	"fmt"

	"github.com/spf13/cobra"

	"vextpss/source/cmd/collectors"
	"vextpss/source/cmd/formatters"
	"vextpss/source/core"
	"vextpss/source/funcs"
	"vextpss/source/secrets"
	"vextpss/source/shared/passgen"
	"vextpss/source/shared/sentinel"
	"vextpss/source/shared/storage"
)

// ================================
// AddCmd returns the cobra command for "vext add <name>".
// enc and c are injected — DB is opened per execution since it may not exist at startup.
// ================================
func AddCmd(dbPath string, enc core.Encryptor, c collectors.Collector) *cobra.Command {
	var secretType string

	cmd := &cobra.Command{
		Use:   "add <name>",
		Short: "Store a new secret",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runAdd(args[0], secretType, dbPath, enc, c)
		},
	}

	cmd.Flags().StringVarP(&secretType, "type", "t", secrets.TypeAccount, `secret type: "account" or "finance"`)
	return cmd
}

func runAdd(name, secretType, dbPath string, enc core.Encryptor, c collectors.Collector) error {
	plaintext, err := c.CollectPayload(secretType)
	defer passgen.Cleaner(plaintext)
	if err != nil {
		formatters.Error(err.Error())
		return err
	}

	masterPassword, err := c.CollectMaster()
	defer passgen.Cleaner(masterPassword)
	if err != nil {
		formatters.Error(err.Error())
		return err
	}

	db, err := storage.Open(dbPath)
	if err != nil {
		formatters.Error(err.Error())
		return err
	}
	defer storage.Close(db)

	f := funcs.NewCreateSecretFunc(storage.NewSecretRepository(db), enc)
	if err = f.Run(context.Background(), funcs.CreateSecretDto{
		Name:           name,
		Type:           secretType,
		Plaintext:      plaintext,
		MasterPassword: masterPassword,
	}); err != nil {
		if errors.Is(err, sentinel.ErrAlreadyExists) {
			formatters.Error(fmt.Sprintf("a credential named %q already exists. Use `vext update` to modify it.", name))
		} else {
			formatters.Error(err.Error())
		}
		return err
	}

	formatters.Success(fmt.Sprintf("Credential %q saved.", name))
	return nil
}
