package addsecret

import (
	"context"
	"fmt"

	"vextpss/source/funcs"
	"vextpss/source/secrets"
	"vextpss/source/shared/memory"
	"vextpss/source/shared/terminal"
)

func run(ctx context.Context, name, secretType string, deps funcs.Deps) error {
	if !secrets.IsKnownType(secretType) {
		return fmt.Errorf("unknown secret type %q — valid types: account, finance", secretType)
	}

	plaintext, master, err := collect(secretType, deps.Prompter)
	if err != nil {
		return err
	}
	defer memory.Cleaner(plaintext)
	defer memory.Cleaner(master)

	encrypted, err := deps.Cryp.Encrypt(ctx, plaintext, master)
	if err != nil {
		return err
	}

	if err := deps.Repo.Create(ctx, secrets.Secret{
		Name:      name,
		Type:      secretType,
		Encrypted: encrypted,
	}); err != nil {
		return err
	}

	terminal.Success(fmt.Sprintf("Secret %q saved.", name))
	return nil
}
