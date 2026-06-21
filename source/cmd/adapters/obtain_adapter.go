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

func GetCmd(app *App) *cobra.Command {
	return &cobra.Command{
		Use:   "get <name>",
		Short: "Retrieve and display a stored secret",
		Args:  cobra.ExactArgs(1),
		RunE: func(_ *cobra.Command, args []string) error {
			return runGet(args[0], *app)
		},
	}
}

func runGet(name string, app App) error {
	ctx := context.Background()

	masterPassword, err := app.Collector.Master()
	defer memory.Cleaner(masterPassword)
	if err != nil {
		formatters.Error(err.Error())
		return err
	}

	result, err := funcs.NewObtainSecretFunc(app.Repository, app.Encryptor).Run(ctx, funcs.ObtainSecretDto{
		Name:           name,
		MasterPassword: masterPassword,
	})
	defer memory.Cleaner(result.Payload)
	if err != nil {
		if errors.Is(err, secrets.ErrSecretNotFound) {
			formatters.Error(fmt.Sprintf("no secret named %q found", name))
		} else if errors.Is(err, secrets.ErrDecryptionFailed) {
			formatters.Error("wrong master password or data corrupted")
		} else {
			formatters.Error(err.Error())
		}
		return err
	}

	formatters.PrintSecret(result.Name, result.Type, result.Payload)
	return nil
}
