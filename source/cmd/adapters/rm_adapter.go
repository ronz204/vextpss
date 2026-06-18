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
	"vextpss/source/shared/sentinel"
	"vextpss/source/shared/storage"
)

// ================================
// RmCmd returns the cobra command for "vext rm <name>".
// No encryptor needed — delete does not touch secret payloads.
// ================================
func RmCmd(deps shared.AppDeps) *cobra.Command {
	return &cobra.Command{
		Use:   "rm <name>",
		Short: "Permanently delete a stored secret",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runRm(args[0], deps)
		},
	}
}

func runRm(name string, deps shared.AppDeps) error {
	ok, err := deps.Collector.Confirm(fmt.Sprintf("Delete %q? This cannot be undone", name))
	if err != nil {
		formatters.Error(err.Error())
		return err
	}
	if !ok {
		formatters.Info("Aborted.")
		return nil
	}

	err = storage.WithRepo(deps.DBPath, func(repo core.SecretRepository) error {
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
