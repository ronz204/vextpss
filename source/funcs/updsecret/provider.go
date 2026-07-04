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

	plaintext, master, err := collect(existing.Type, deps.Prompter)
	if err != nil {
		return err
	}
	defer memory.Cleaner(plaintext)
	defer memory.Cleaner(master)

	encrypted, err := deps.Cryp.Encrypt(ctx, plaintext, master)
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
