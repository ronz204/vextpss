package listsecrets

import (
	"context"

	"vextpss/source/funcs"
)

func run(ctx context.Context, deps funcs.Deps) error {
	list, err := deps.SecretRepo.List(ctx, deps.ActiveSpace)
	if err != nil {
		return err
	}

	printSecretsTable(list)
	return nil
}
