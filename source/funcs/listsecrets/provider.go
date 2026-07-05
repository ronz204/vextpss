package listsecrets

import (
	"context"

	"vextpss/source/funcs"
)

func run(ctx context.Context, deps funcs.Deps) error {
	list, err := deps.Repo.List(ctx)
	if err != nil {
		return err
	}

	printSecretsTable(list)
	return nil
}
