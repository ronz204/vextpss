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
	"vextpss/source/shared/passgen"
	"vextpss/source/shared/sentinel"
	"vextpss/source/shared/storage"
)

// ================================
// UpdCmd returns the cobra command for "vext upd <name>".
// The secret type is resolved from the existing record — no --type flag required.
// ================================
func UpdCmd(dbPath string, enc core.Encryptor, c collectors.Collector) *cobra.Command {
	return &cobra.Command{
		Use:   "upd <name>",
		Short: "Update an existing secret",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runUpd(args[0], dbPath, enc, c)
		},
	}
}

func runUpd(name, dbPath string, enc core.Encryptor, c collectors.Collector) error {
	db, err := storage.Open(dbPath)
	if err != nil {
		formatters.Error(err.Error())
		return err
	}
	defer storage.Close(db)

	repo := storage.NewSecretRepository(db)

	// Resolve the existing secret type before prompting — the user should not need to remember it.
	existing, _, err := repo.GetByName(context.Background(), name)
	if err != nil {
		if errors.Is(err, sentinel.ErrSecretNotFound) {
			formatters.Error(fmt.Sprintf("no secret named %q found", name))
		} else {
			formatters.Error(err.Error())
		}
		return err
	}

	plaintext, err := c.CollectPayload(existing.Type)
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

	f := funcs.NewUpdateSecretFunc(repo, enc)
	if err = f.Run(context.Background(), funcs.UpdateSecretDto{
		Name:           name,
		Type:           existing.Type,
		Plaintext:      plaintext,
		MasterPassword: masterPassword,
	}); err != nil {
		if errors.Is(err, sentinel.ErrSecretNotFound) {
			formatters.Error(fmt.Sprintf("no secret named %q found", name))
		} else {
			formatters.Error(err.Error())
		}
		return err
	}

	formatters.Success(fmt.Sprintf("Secret %q updated.", name))
	return nil
}
