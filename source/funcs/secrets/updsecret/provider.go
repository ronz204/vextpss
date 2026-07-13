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
	existing, err := deps.SecretRepo.GetByName(ctx, deps.State.ActiveSpace, name)
	if err != nil {
		return err
	}

	master, err := deps.Prompter.ReadSecret("Master password")
	if err != nil {
		return err
	}
	defer memory.Cleaner(master)

	currentPlaintext, err := deps.Cypher.Decrypt(ctx, existing.Encrypted, master)
	if err != nil {
		return err
	}
	defer memory.Cleaner(currentPlaintext)

	newPlaintext, err := collect(existing.Type, currentPlaintext, deps.Prompter)
	if err != nil {
		return err
	}
	defer memory.Cleaner(newPlaintext)

	encrypted, err := deps.Cypher.Encrypt(ctx, newPlaintext, master)
	if err != nil {
		return err
	}

	if err := deps.SecretRepo.Update(ctx, core.Secret{
		Space:     deps.State.ActiveSpace,
		Name:      name,
		Type:      existing.Type,
		Encrypted: encrypted,
	}); err != nil {
		return err
	}

	terminal.Success(fmt.Sprintf("Secret %q updated.", name))
	return nil
}
