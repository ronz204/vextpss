package main

import (
	"os"

	"github.com/spf13/cobra"

	"vextpss/source/funcs"
	"vextpss/source/funcs/secrets/addsecret"
	"vextpss/source/funcs/secrets/dropsecret"
	"vextpss/source/funcs/secrets/gensecret"
	"vextpss/source/funcs/secrets/getsecrets"
	"vextpss/source/funcs/secrets/listsecrets"
	"vextpss/source/funcs/secrets/rensecret"
	"vextpss/source/funcs/secrets/rotasecret"
	"vextpss/source/funcs/secrets/updsecret"
	"vextpss/source/shared/cyphers/aesgcm"
	"vextpss/source/shared/storages"
	"vextpss/source/shared/terminal"
)

func loadDeps() (funcs.Deps, error) {
	dbPath, err := storages.DefaultDBPath()
	if err != nil {
		return funcs.Deps{}, err
	}
	db, state, err := storages.Init(dbPath)
	if err != nil {
		return funcs.Deps{}, err
	}
	return funcs.Deps{
		SecretRepo: storages.NewSecrets(db),
		SpaceRepo:  storages.NewSpaces(db),
		StateRepo:  storages.NewState(db),
		Cypher:     aesgcm.New(aesgcm.DefaultConfig()),
		Prompter:   terminal.NewPrompter(),
		State:      state,
	}, nil
}

func main() {
	deps, err := loadDeps()
	if err != nil {
		terminal.Error(err.Error())
		os.Exit(1)
	}

	root := &cobra.Command{
		Use:           "vext",
		Short:         "A local, encrypted password manager",
		SilenceUsage:  true,
		SilenceErrors: true,
	}

	root.AddCommand(addsecret.NewCommand(deps))
	root.AddCommand(gensecret.NewCommand())
	root.AddCommand(dropsecret.NewCommand(deps))
	root.AddCommand(getsecrets.NewCommand(deps))
	root.AddCommand(listsecrets.NewCommand(deps))
	root.AddCommand(rensecret.NewCommand(deps))
	root.AddCommand(rotasecret.NewCommand(deps))
	root.AddCommand(updsecret.NewCommand(deps))

	if err := root.Execute(); err != nil {
		terminal.Error(err.Error())
		os.Exit(1)
	}
}
