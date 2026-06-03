package commands

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"strings"

	"vextpss/source/pkg/database"

	"github.com/spf13/cobra"
)

var rmCmd = &cobra.Command{
	Use:   "rm <name>",
	Short: "Delete a stored credential",
	Long:  "Permanently removes a stored credential by name after a confirmation prompt.",
	Args:  cobra.ExactArgs(1),
	RunE:  runRm,
}

func runRm(cmd *cobra.Command, args []string) error {
	name := args[0]

	// Confirmation prompt — destructive operation.
	fmt.Printf("Are you sure you want to delete %q? This cannot be undone. [y/N]: ", name)
	reader := bufio.NewReader(os.Stdin)
	answer, err := reader.ReadString('\n')
	if err != nil {
		return fmt.Errorf("could not read confirmation: %w", err)
	}
	answer = strings.TrimSpace(strings.ToLower(answer))
	if answer != "y" && answer != "yes" {
		fmt.Println("Aborted.")
		return nil
	}

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

	if err := database.DeleteByName(db, name); errors.Is(err, database.ErrNotFound) {
		return fmt.Errorf("[X] Error: no credential named %q found", name)
	} else if err != nil {
		return err
	}

	fmt.Printf("[✓] Credential %q deleted.\n", name)
	return nil
}
