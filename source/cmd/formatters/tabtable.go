package formatters

import (
	"fmt"
	"os"
	"text/tabwriter"

	"vextpss/source/secrets"
)

// PrintTabTable prints a formatted table of secret names, types, and creation dates.
func PrintTabTable(ss []secrets.Secret) {
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 3, ' ', 0)
	fmt.Fprintln(w, "NAME\tTYPE\tCREATED")
	for _, s := range ss {
		fmt.Fprintf(w, "%s\t%s\t%s\n", s.Name, s.Type, s.CreatedAt.Format("2006-01-02"))
	}
	w.Flush()
}
