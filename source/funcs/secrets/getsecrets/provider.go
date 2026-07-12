package getsecrets

import (
	"context"

	"vextpss/source/funcs"
	"vextpss/source/secrets/moon"
	"vextpss/source/shared/memory"
)

func run(ctx context.Context, name string, deps funcs.Deps) error {
	secret, err := deps.SecretRepo.GetByName(ctx, deps.ActiveSpace, name)
	if err != nil {
		return err
	}

	master, err := deps.Prompter.ReadSecret("Master password")
	if err != nil {
		return err
	}
	defer memory.Cleaner(master)

	plaintext, err := deps.Cypher.Decrypt(ctx, secret.Encrypted, master)
	if err != nil {
		return err
	}
	defer memory.Cleaner(plaintext)

	p, err := moon.Decode(secret.Type, plaintext)
	if err != nil {
		return err
	}
	p.Display()
	return nil
}
