package dropsecret

import (
	"context"
	"fmt"

	"vextpss/source/funcs"
	"vextpss/source/shared/terminal"
)

func run(ctx context.Context, name string, deps funcs.Deps) error {
	ok, err := deps.Prompter.Confirm(fmt.Sprintf("Delete secret %q", name))
	if err != nil {
		return err
	}
	if !ok {
		return nil
	}

	if err := deps.Repo.Delete(ctx, name); err != nil {
		return err
	}

	terminal.Success(fmt.Sprintf("Secret %q deleted.", name))
	return nil
}
