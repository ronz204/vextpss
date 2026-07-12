package listsecrets

import (
	"fmt"
	"os"
	"text/tabwriter"

	"vextpss/source/secrets/core"
	"vextpss/source/shared/terminal"
)

func printSecretsTable(list []core.Secret) {
	if len(list) == 0 {
		terminal.Info("No secrets stored yet.")
		return
	}
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 3, ' ', 0)
	fmt.Fprintln(w, "  NAME\tTYPE\tCREATED")
	for _, s := range list {
		fmt.Fprintf(w, "  %s\t%s\t%s\n", s.Name, s.Type, s.CreatedAt.Format("2006-01-02"))
	}
	w.Flush()
}
