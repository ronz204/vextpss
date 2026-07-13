package dropspace

import (
	"context"
	"errors"
	"fmt"

	"vextpss/source/funcs"
	"vextpss/source/shared/terminal"
)

func run(ctx context.Context, name string, force bool, deps funcs.Deps) error {
	if name == deps.State.ActiveSpace {
		return errors.New("cannot drop the active space — switch with 'vext use <other>' first")
	}

	secrets, err := deps.SecretRepo.List(ctx, name)
	if err != nil {
		return err
	}

	if len(secrets) > 0 && !force {
		return fmt.Errorf("space %q has %d secret(s) — use --force to delete them all", name, len(secrets))
	}

	if len(secrets) > 0 {
		ok, err := deps.Prompter.Confirm(fmt.Sprintf("Delete space %q and all %d secret(s) inside?", name, len(secrets)))
		if err != nil {
			return err
		}
		if !ok {
			return nil
		}
		for _, s := range secrets {
			if err := deps.SecretRepo.Delete(ctx, name, s.Name); err != nil {
				return err
			}
		}
	} else {
		ok, err := deps.Prompter.Confirm(fmt.Sprintf("Delete space %q?", name))
		if err != nil {
			return err
		}
		if !ok {
			return nil
		}
	}

	if err := deps.SpaceRepo.Delete(ctx, name); err != nil {
		return err
	}

	terminal.Success(fmt.Sprintf("Space %q deleted.", name))
	return nil
}
