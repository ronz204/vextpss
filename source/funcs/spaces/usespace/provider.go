package usespace

import (
	"context"
	"fmt"

	"vextpss/source/funcs"
	"vextpss/source/shared/terminal"
)

func showActive(deps funcs.Deps) error {
	terminal.Info(fmt.Sprintf("Active space: %s", deps.State.ActiveSpace))
	return nil
}

func run(ctx context.Context, name string, deps funcs.Deps) error {
	if _, err := deps.SpaceRepo.GetByName(ctx, name); err != nil {
		return err
	}
	newState := deps.State
	newState.ActiveSpace = name
	if err := deps.StateRepo.Save(ctx, newState); err != nil {
		return err
	}
	terminal.Success(fmt.Sprintf("Now using space %q.", name))
	return nil
}
