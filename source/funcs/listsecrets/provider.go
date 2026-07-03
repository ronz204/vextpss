package listsecrets

import (
	"context"

	"vextpss/source/funcs"
	"vextpss/source/shared/terminal"
)

func run(ctx context.Context, deps funcs.Deps) error {
	list, err := deps.Repo.List(ctx)
	if err != nil {
		return err
	}

	terminal.PrintSecretsTable(list)
	return nil
}
