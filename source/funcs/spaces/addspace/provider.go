package addspace

import (
	"context"
	"fmt"

	"vextpss/source/funcs"
	"vextpss/source/shared/terminal"
)

func run(ctx context.Context, name string, deps funcs.Deps) error {
	if err := deps.SpaceRepo.Create(ctx, name); err != nil {
		return err
	}
	terminal.Success(fmt.Sprintf("Space %q created.", name))
	return nil
}
