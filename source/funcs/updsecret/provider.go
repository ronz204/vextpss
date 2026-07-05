package updsecret

import (
	"context"
	"fmt"

	"vextpss/source/funcs"
	"vextpss/source/secrets/core"
	"vextpss/source/shared/memory"
	"vextpss/source/shared/terminal"
)

func run(ctx context.Context, name string, deps funcs.Deps) error {
	existing, err := deps.Repo.GetByName(ctx, name)
	if err != nil {
		return err
	}

	master, err := deps.Prompter.ReadSecret("Master password")
	if err != nil {
		return err
	}
	defer memory.Cleaner(master)

	currentPlaintext, err := deps.Cryp.Decrypt(ctx, existing.Encrypted, master)
	if err != nil {
		return err
	}
	defer memory.Cleaner(currentPlaintext)

	newPlaintext, err := collect(existing.Type, currentPlaintext, deps.Prompter)
	if err != nil {
		return err
	}
	defer memory.Cleaner(newPlaintext)

	encrypted, err := deps.Cryp.Encrypt(ctx, newPlaintext, master)
	if err != nil {
		return err
	}

	if err := deps.Repo.Update(ctx, core.Secret{
		Name:      name,
		Type:      existing.Type,
		Encrypted: encrypted,
	}); err != nil {
		return err
	}

	terminal.Success(fmt.Sprintf("Secret %q updated.", name))
	return nil
}
