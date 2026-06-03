package commands

import (
	"fmt"

	"vextpss/source/pkg/database"

	"github.com/spf13/cobra"
)

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List all stored secrets",
	Long:  "Displays a table of all stored secret names and types. Does not require the master password.",
	Args:  cobra.NoArgs,
	RunE:  runList,
}

func runList(cmd *cobra.Command, args []string) error {
	dbPath, err := database.DBPath()
	if err != nil {
		return err
	}
	db, err := database.Open(dbPath)
	if err != nil {
		return err
	}
	sqlDB, err := db.DB()
	if err != nil {
		return fmt.Errorf("could not access underlying db: %w", err)
	}
	defer sqlDB.Close()

	records, err := database.ListAll(db)
	if err != nil {
		return err
	}

	if len(records) == 0 {
		fmt.Println("No secrets stored. Use `vext add <name>` to add one.")
		return nil
	}

	fmt.Printf("%-30s  %-12s  %s\n", "NAME", "TYPE", "CREATED")
	fmt.Printf("%-30s  %-12s  %s\n", "------------------------------", "------------", "-------------------")
	for _, r := range records {
		fmt.Printf("%-30s  %-12s  %s\n", r.Name, r.Type, r.CreatedAt.Format("2006-01-02 15:04:05"))
	}

	return nil
}
