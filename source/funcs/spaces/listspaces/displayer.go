package listspaces

import (
	"fmt"
	"os"
	"text/tabwriter"

	"vextpss/source/secrets/core"
	"vextpss/source/shared/terminal"
)

func printSpacesTable(spaces []core.Space, counts map[string]int, active string) {
	if len(spaces) == 0 {
		terminal.Info("No spaces yet. Run 'vext spaces add <name>'.")
		return
	}
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 3, ' ', 0)
	fmt.Fprintln(w, "  SPACE\tSECRETS\tACTIVE")
	for _, sp := range spaces {
		marker := ""
		if sp.Name == active {
			marker = "*"
		}
		fmt.Fprintf(w, "  %s\t%d\t%s\n", sp.Name, counts[sp.Name], marker)
	}
	w.Flush()
}
