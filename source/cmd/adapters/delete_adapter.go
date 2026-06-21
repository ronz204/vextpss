package adapters

import (
	"context"
	"errors"
	"fmt"

	"github.com/spf13/cobra"

	"vextpss/source/cmd/formatters"
	"vextpss/source/funcs"
	"vextpss/source/secrets"
)

func RmCmd(app App) *cobra.Command {
	return &cobra.Command{
		Use:   "rm <name>",
		Short: "Permanently delete a stored secret",
		Args:  cobra.ExactArgs(1),
		RunE: func(_ *cobra.Command, args []string) error {
			return runRm(args[0], app)
		},
	}
}

func runRm(name string, app App) error {
	ctx := context.Background()

	ok, err := app.Collector.Confirm(fmt.Sprintf("Delete %q? This cannot be undone", name))
	if err != nil {
		formatters.Error(err.Error())
		return err
	}
	if !ok {
		fmt.Println("Aborted.")
		return nil
	}

	if err := funcs.NewDeleteSecretFunc(app.Repository).Run(ctx, funcs.DeleteSecretDto{Name: name}); err != nil {
		if errors.Is(err, secrets.ErrSecretNotFound) {
			formatters.Error(fmt.Sprintf("no secret named %q found", name))
		} else {
			formatters.Error(err.Error())
		}
		return err
	}

	formatters.Success(fmt.Sprintf("Secret %q deleted.", name))
	return nil
}
