package renspace

import (
	"context"
	"fmt"

	"vextpss/source/funcs"
	"vextpss/source/shared/terminal"
)

func run(ctx context.Context, oldName, newName string, deps funcs.Deps) error {
	if err := deps.SpaceRepo.Rename(ctx, oldName, newName); err != nil {
		return err
	}

	if deps.State.ActiveSpace == oldName {
		newState := deps.State
		newState.ActiveSpace = newName
		if err := deps.StateRepo.Save(ctx, newState); err != nil {
			return err
		}
	}

	terminal.Success(fmt.Sprintf("Space %q renamed to %q.", oldName, newName))
	return nil
}
