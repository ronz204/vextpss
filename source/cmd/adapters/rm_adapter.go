package adapters

import (
	"context"
	"errors"
	"fmt"

	"github.com/spf13/cobra"

	"vextpss/source/cmd/collectors"
	"vextpss/source/cmd/formatters"
	"vextpss/source/funcs"
	"vextpss/source/shared/sentinel"
	"vextpss/source/shared/storage"
)

// ================================
// RmCmd returns the cobra command for "vext rm <name>".
// No encryptor needed — delete does not touch secret payloads.
// ================================
func RmCmd(dbPath string, c collectors.Collector) *cobra.Command {
	return &cobra.Command{
		Use:   "rm <name>",
		Short: "Permanently delete a stored secret",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runRm(args[0], dbPath, c)
		},
	}
}

func runRm(name, dbPath string, c collectors.Collector) error {
	ok, err := c.Confirm(fmt.Sprintf("Delete %q? This cannot be undone", name))
	if err != nil {
		formatters.Error(err.Error())
		return err
	}
	if !ok {
		formatters.Info("Aborted.")
		return nil
	}

	db, err := storage.Open(dbPath)
	if err != nil {
		formatters.Error(err.Error())
		return err
	}
	defer storage.Close(db)

	f := funcs.NewDeleteSecretFunc(storage.NewSecretRepository(db))
	if err = f.Run(context.Background(), funcs.DeleteSecretDto{Name: name}); err != nil {
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
