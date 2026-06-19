package adapters

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/spf13/cobra"

	"vextpss/source/cmd/formatters"
	"vextpss/source/funcs"
	"vextpss/source/secrets"
	"vextpss/source/shared/memory"
	"vextpss/source/shared/sentinel"
	"vextpss/source/shared/storage"
)

func GetCmd(d Deps) *cobra.Command {
	return &cobra.Command{
		Use:   "get <name>",
		Short: "Retrieve and decrypt a stored secret",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runGet(args[0], d)
		},
	}
}

func runGet(name string, d Deps) error {
	masterPassword, err := d.Collector.Master()
	defer memory.Cleaner(masterPassword)
	if err != nil {
		formatters.Error(err.Error())
		return err
	}

	var result funcs.ObtainSecretResult
	err = storage.WithRepo(d.DBPath, func(repo funcs.Repository) error {
		var runErr error
		result, runErr = funcs.NewObtainSecretFunc(repo, d.Encryptor).Run(context.Background(), funcs.ObtainSecretDto{
			Name:           name,
			MasterPassword: masterPassword,
		})
		return runErr
	})
	defer memory.Cleaner(result.Payload)
	if err != nil {
		if errors.Is(err, sentinel.ErrSecretNotFound) {
			formatters.Error(fmt.Sprintf("no secret named %q found", name))
		} else if errors.Is(err, sentinel.ErrDecryptionFailed) {
			formatters.Error("wrong master password or data corrupted")
		} else {
			formatters.Error(err.Error())
		}
		return err
	}

	return printSecret(result)
}

func printSecret(result funcs.ObtainSecretResult) error {
	switch result.Type {
	case secrets.TypeAccount:
		var s secrets.AccountSecret
		if err := json.Unmarshal(result.Payload, &s); err != nil {
			return fmt.Errorf("decode account secret: %w", err)
		}
		formatters.PrintAccount(result.Name, s)
	case secrets.TypeFinance:
		var s secrets.FinanceSecret
		if err := json.Unmarshal(result.Payload, &s); err != nil {
			return fmt.Errorf("decode finance secret: %w", err)
		}
		formatters.PrintFinance(result.Name, s)
	default:
		return fmt.Errorf("unknown secret type %q", result.Type)
	}
	return nil
}
