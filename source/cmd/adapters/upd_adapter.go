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
	"vextpss/source/shared/passgen"
	"vextpss/source/shared/sentinel"
	"vextpss/source/shared/storage"
)

// ================================
// UpdCmd returns the cobra command for "vext upd <name>".
// The secret type is resolved from the existing record — no --type flag required.
// ================================
func UpdCmd(deps shared.AppDeps) *cobra.Command {
	return &cobra.Command{
		Use:   "upd <name>",
		Short: "Update an existing secret",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runUpd(args[0], deps)
		},
	}
}

func runUpd(name string, deps shared.AppDeps) error {
	err := storage.WithRepo(deps.DBPath, func(repo core.SecretRepository) error {
		// Resolve the existing type before prompting — the user should not need to remember it.
		existing, _, err := repo.GetByName(context.Background(), name)
		if err != nil {
			return err
		}

		plaintext, err := deps.Collector.CollectPayload(existing.Type)
		defer passgen.Cleaner(plaintext)
		if err != nil {
			return err
		}

		masterPassword, err := deps.Collector.CollectMaster()
		defer passgen.Cleaner(masterPassword)
		if err != nil {
			return err
		}

		return funcs.NewUpdateSecretFunc(repo, deps.Enc).Run(context.Background(), funcs.UpdateSecretDto{
			Name:           name,
			Type:           existing.Type,
			Plaintext:      plaintext,
			MasterPassword: masterPassword,
		})
	})

	if err != nil {
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
