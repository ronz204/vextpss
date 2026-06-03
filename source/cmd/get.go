package commands

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"

	"vextpss/source/pkg/crypto"
	"vextpss/source/pkg/database"
	"vextpss/source/pkg/models"

	"github.com/spf13/cobra"
	"golang.org/x/term"
)

var getCmd = &cobra.Command{
	Use:   "get <name>",
	Short: "Retrieve and display a stored credential",
	Long:  "Looks up a stored credential by name, decrypts it, and displays its fields.",
	Args:  cobra.ExactArgs(1),
	RunE:  runGet,
}

func runGet(cmd *cobra.Command, args []string) error {
	name := args[0]

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

	record, err := database.GetByName(db, name)
	if errors.Is(err, database.ErrNotFound) {
		return fmt.Errorf("[X] Error: no credential named %q found", name)
	}
	if err != nil {
		return err
	}

	// Collect master password (hidden).
	fmt.Print("Master Password: ")
	masterPasswordBytes, err := term.ReadPassword(int(os.Stdin.Fd()))
	fmt.Println()
	if err != nil {
		return fmt.Errorf("could not read master password: %w", err)
	}
	defer crypto.Zero(masterPasswordBytes)

	// Derive the key using the per-record salt stored in the database.
	key := crypto.DeriveKey(masterPasswordBytes, record.Salt)
	defer crypto.Zero(key)

	// Decrypt — returns ErrDecryptFailed on wrong password or tampered data.
	plaintextJSON, err := crypto.Decrypt(key, record.Nonce, record.EncryptedPayload)
	if err != nil {
		return fmt.Errorf("[X] Error: %s", err)
	}
	defer crypto.Zero(plaintextJSON)

	// Dispatch to the correct payload type based on the type tag in the record.
	return displaySecret(record.Type, record.Name, plaintextJSON)
}

// displaySecret deserializes and prints the decrypted payload for the given secret type.
func displaySecret(secretType, name string, plaintextJSON []byte) error {
	switch secretType {
	case "account":
		var payload models.AccountPayload
		if err := json.Unmarshal(plaintextJSON, &payload); err != nil {
			return fmt.Errorf("could not parse credential: %w", err)
		}
		fmt.Printf("Service:  %s\n", name)
		fmt.Printf("Username: %s\n", payload.Username)
		fmt.Printf("Password: %s\n", payload.Password)
		// Zero out the password field after printing.
		crypto.Zero([]byte(payload.Password))
	default:
		return fmt.Errorf("unknown secret type %q", secretType)
	}
	return nil
}
