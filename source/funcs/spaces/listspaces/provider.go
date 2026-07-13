package listspaces

import (
	"context"

	"vextpss/source/funcs"
)

func Run(ctx context.Context, deps funcs.Deps) error {
	spaces, err := deps.SpaceRepo.List(ctx)
	if err != nil {
		return err
	}

	allSecrets, err := deps.SecretRepo.List(ctx, "")
	if err != nil {
		return err
	}

	counts := make(map[string]int, len(spaces))
	for _, s := range allSecrets {
		counts[s.Space]++
	}

	printSpacesTable(spaces, counts, deps.State.ActiveSpace)
	return nil
}
