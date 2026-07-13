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
	existing, err := deps.SecretRepo.GetByName(ctx, deps.State.ActiveSpace, name)
	if err != nil {
		return err
	}

	currentMaster, err := deps.Prompter.ReadSecret("Current master password")
	if err != nil {
		return err
	}
	defer memory.Cleaner(currentMaster)

	plaintext, err := deps.Cypher.Decrypt(ctx, existing.Encrypted, currentMaster)
	if err != nil {
		return err
	}
	defer memory.Cleaner(plaintext)

	newMaster, err := deps.Prompter.ReadSecret("New master password")
	if err != nil {
		return err
	}
	defer memory.Cleaner(newMaster)

	encrypted, err := deps.Cypher.Encrypt(ctx, plaintext, newMaster)
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

	terminal.Success(fmt.Sprintf("Master password for %q rotated.", name))
	return nil
}
