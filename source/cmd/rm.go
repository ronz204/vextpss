package cmd

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

func RmCmd() *cobra.Command {
	c := &cobra.Command{
		Use:   "rm <name>",
		Short: "Permanently delete a stored secret",
		Args:  cobra.ExactArgs(1),
	}
	c.RunE = withDeps(func(ctx context.Context, d Deps, args []string) error {
		return runRm(ctx, d, args[0])
	})
	return c
}

func runRm(ctx context.Context, d Deps, name string) error {
	ok, err := d.Collect.Confirm(fmt.Sprintf("Delete %q? This cannot be undone", name))
	if err != nil {
		formatters.Error(err.Error())
		return err
	}
	if !ok {
		fmt.Println("Aborted.")
		return nil
	}

	masterPassword, err := d.Collect.Master()
	defer memory.Cleaner(masterPassword)
	if err != nil {
		formatters.Error(err.Error())
		return err
	}

	if err := funcs.NewDeleteSecretFunc(d.Repo, d.Enc).Run(ctx, funcs.DeleteSecretDto{
		Name:           name,
		MasterPassword: masterPassword,
	}); err != nil {
		if errors.Is(err, secrets.ErrSecretNotFound) {
			formatters.Error(fmt.Sprintf("no secret named %q found", name))
		} else if errors.Is(err, secrets.ErrDecryptionFailed) {
			formatters.Error("wrong master password — deletion aborted")
		} else {
			formatters.Error(err.Error())
		}
		return err
	}

	formatters.Success(fmt.Sprintf("Secret %q deleted.", name))
	return nil
}
