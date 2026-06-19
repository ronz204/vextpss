package adapters

import (
	"context"
	"errors"
	"fmt"

	"github.com/spf13/cobra"

	"vextpss/source/cmd/formatters"
	"vextpss/source/funcs"
	"vextpss/source/shared/sentinel"
	"vextpss/source/shared/storage"
)

func RmCmd(d Deps) *cobra.Command {
	return &cobra.Command{
		Use:   "rm <name>",
		Short: "Permanently delete a stored secret",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runRm(args[0], d)
		},
	}
}

func runRm(name string, d Deps) error {
	ok, err := d.Collector.Confirm(fmt.Sprintf("Delete %q? This cannot be undone", name))
	if err != nil {
		formatters.Error(err.Error())
		return err
	}
	if !ok {
		formatters.Info("Aborted.")
		return nil
	}

	err = storage.WithRepo(d.DBPath, func(repo funcs.Repository) error {
		return funcs.NewDeleteSecretFunc(repo).Run(context.Background(), funcs.DeleteSecretDto{Name: name})
	})

	if err != nil {
		if errors.Is(err, sentinel.ErrSecretNotFound) {
			formatters.Error(fmt.Sprintf("no secret named %q found", name))
		} else {
			formatters.Error(err.Error())
		}
		return err
	}

	formatters.Success(fmt.Sprintf("Secret %q deleted.", name))
	return nil
}
