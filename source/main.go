package main

import (
	"os"

	"github.com/spf13/cobra"

	"vextpss/source/funcs"
	"vextpss/source/funcs/addsecret"
	"vextpss/source/funcs/dropsecret"
	"vextpss/source/funcs/getsecrets"
	"vextpss/source/funcs/listsecrets"
	"vextpss/source/funcs/updsecret"
	"vextpss/source/shared/cyphers/aesgcm"
	"vextpss/source/shared/storages"
	"vextpss/source/shared/terminal"
)

func loadDeps() (funcs.Deps, error) {
	dbPath, err := storages.DefaultDBPath()
	if err != nil {
		return funcs.Deps{}, err
	}
	db, err := storages.Init(dbPath)
	if err != nil {
		return funcs.Deps{}, err
	}
	return funcs.Deps{
		Repo:     storages.New(db),
		Cryp:     aesgcm.New(aesgcm.DefaultConfig()),
		Prompter: terminal.NewPrompter(),
	}, nil
}

func main() {
	deps, err := loadDeps()
	if err != nil {
		terminal.Error(err.Error())
		os.Exit(1)
	}

	root := &cobra.Command{
		Use:          "vext",
		Short:        "A local, encrypted password manager",
		SilenceUsage: true,
	}

	root.AddCommand(addsecret.NewCommand(deps))
	root.AddCommand(dropsecret.NewCommand(deps))
	root.AddCommand(getsecrets.NewCommand(deps))
	root.AddCommand(listsecrets.NewCommand(deps))
	root.AddCommand(updsecret.NewCommand(deps))

	if err := root.Execute(); err != nil {
		os.Exit(1)
	}
}
