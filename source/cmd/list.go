package cmd

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"

	"vextpss/source/cmd/formatters"
	"vextpss/source/funcs"
)

func ListCmd() *cobra.Command {
	c := &cobra.Command{
		Use:   "list",
		Short: "List all stored secrets",
		Args:  cobra.NoArgs,
	}
	c.RunE = withDeps(func(ctx context.Context, d Deps, _ []string) error {
		return runList(ctx, d)
	})
	return c
}

func runList(ctx context.Context, d Deps) error {
	all, err := funcs.NewRetrieveSecretsFunc(d.Repo).Run(ctx)
	if err != nil {
		formatters.Error(err.Error())
		return err
	}

	formatters.PrintTabTable(all)
	fmt.Printf("Total: %d secrets.\n", len(all))
	return nil
}
