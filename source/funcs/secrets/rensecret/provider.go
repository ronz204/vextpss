package rensecret

import (
	"context"
	"fmt"

	"vextpss/source/funcs"
	"vextpss/source/shared/terminal"
)

func run(ctx context.Context, oldName, newName string, deps funcs.Deps) error {
	if err := deps.SecretRepo.Rename(ctx, deps.ActiveSpace, oldName, newName); err != nil {
		return err
	}

	terminal.Success(fmt.Sprintf("Secret %q renamed to %q.", oldName, newName))
	return nil
}
