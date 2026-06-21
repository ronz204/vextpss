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
)

func AddCmd(app *App) *cobra.Command {
	var secretType string

	c := &cobra.Command{
		Use:   "add <name>",
		Short: "Store a new secret",
		Args:  cobra.ExactArgs(1),
		RunE: func(_ *cobra.Command, args []string) error {
			return runAdd(args[0], secretType, *app)
		},
	}

	c.Flags().StringVarP(&secretType, "type", "t", secrets.TypeAccount, `secret type: "account" or "finance"`)
	return c
}

func runAdd(name, secretType string, app App) error {
	ctx := context.Background()

	plaintext, err := app.Collector.Payload(secretType)
	defer memory.Cleaner(plaintext)
	if err != nil {
		formatters.Error(err.Error())
		return err
	}

	masterPassword, err := app.Collector.Master()
	defer memory.Cleaner(masterPassword)
	if err != nil {
		formatters.Error(err.Error())
		return err
	}

	if err := funcs.NewCreateSecretFunc(app.Repository, app.Encryptor).Run(ctx, funcs.CreateSecretDto{
		Name:           name,
		Type:           secretType,
		Plaintext:      plaintext,
		MasterPassword: masterPassword,
	}); err != nil {
		if errors.Is(err, secrets.ErrAlreadyExists) {
			formatters.Error(fmt.Sprintf("a credential named %q already exists — use `vext upd` to modify it", name))
		} else {
			formatters.Error(err.Error())
		}
		return err
	}

	formatters.Success(fmt.Sprintf("Credential %q saved.", name))
	return nil
}
