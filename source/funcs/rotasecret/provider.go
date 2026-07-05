package rotasecret

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

	currentMaster, err := deps.Prompter.ReadSecret("Current master password")
	if err != nil {
		return err
	}
	defer memory.Cleaner(currentMaster)

	plaintext, err := deps.Cryp.Decrypt(ctx, existing.Encrypted, currentMaster)
	if err != nil {
		return err
	}
	defer memory.Cleaner(plaintext)

	newMaster, err := deps.Prompter.ReadSecret("New master password")
	if err != nil {
		return err
	}
	defer memory.Cleaner(newMaster)

	encrypted, err := deps.Cryp.Encrypt(ctx, plaintext, newMaster)
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

	terminal.Success(fmt.Sprintf("Master password for %q rotated.", name))
	return nil
}
