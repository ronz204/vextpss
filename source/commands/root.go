package commands

import (
	"fmt"

	"github.com/spf13/cobra"
)

var RootCmd = &cobra.Command{
	Use:   "vext",
	Short: "short sample",
	Long:  "long sample",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Hello, from Cobra Vext!")
	},
}

func Execute() {
	if err := RootCmd.Execute(); err != nil {
		fmt.Println(err)
	}
}
