package adapters

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"

	"vextpss/source/cmd/formatters"
	"vextpss/source/funcs"
)

func ListCmd(app *App) *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List all stored secrets",
		Args:  cobra.NoArgs,
		RunE: func(_ *cobra.Command, _ []string) error {
			return runList(*app)
		},
	}
}

func runList(app App) error {
	ctx := context.Background()

	all, err := funcs.NewRetrieveSecretsFunc(app.Repository).Run(ctx)
	if err != nil {
		formatters.Error(err.Error())
		return err
	}

	formatters.PrintTabTable(all)
	fmt.Printf("Total: %d secrets.\n", len(all))
	return nil
}
