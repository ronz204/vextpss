package commands

import (
	"encoding/json"
	"fmt"
	"os"

	"vextpss/source/pkg/crypto"
	"vextpss/source/pkg/database"
	"vextpss/source/pkg/models"

	"github.com/spf13/cobra"
	"golang.org/x/term"
)

var addCmd = &cobra.Command{
	Use:   "add <name>",
	Short: "Store a new credential",
	Long:  "Interactively stores a new account credential under the given name.",
	Args:  cobra.ExactArgs(1),
	RunE:  runAdd,
}

func runAdd(cmd *cobra.Command, args []string) error {
	name := args[0]

	// Collect username (visible — not sensitive).
	fmt.Print("Username: ")
	var username string
	if _, err := fmt.Fscan(os.Stdin, &username); err != nil {
		return fmt.Errorf("could not read username: %w", err)
	}

	// Collect service password (hidden).
	fmt.Print("Password: ")
	servicePasswordBytes, err := term.ReadPassword(int(os.Stdin.Fd()))
	fmt.Println()
	if err != nil {
		return fmt.Errorf("could not read password: %w", err)
	}
	defer crypto.Zero(servicePasswordBytes)

	// Collect master password (hidden).
	fmt.Print("Master Password: ")
	masterPasswordBytes, err := term.ReadPassword(int(os.Stdin.Fd()))
	fmt.Println()
	if err != nil {
		crypto.Zero(servicePasswordBytes)
		return fmt.Errorf("could not read master password: %w", err)
	}
	defer crypto.Zero(masterPasswordBytes)

	// Build the typed payload and serialize to JSON.
	payload := models.AccountPayload{
		Username: username,
		Password: string(servicePasswordBytes),
	}
	plaintextJSON, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("could not serialize payload: %w", err)
	}
	defer crypto.Zero(plaintextJSON)

	// Generate per-record salt and derive encryption key.
	salt, err := crypto.GenerateSalt()
	if err != nil {
		return fmt.Errorf("could not generate salt: %w", err)
	}

	key := crypto.DeriveKey(masterPasswordBytes, salt)
	defer crypto.Zero(key)

	// Encrypt the JSON payload.
	nonce, ciphertext, err := crypto.Encrypt(key, plaintextJSON)
	if err != nil {
		return fmt.Errorf("encryption failed: %w", err)
	}

	// Persist the encrypted record — the database layer never sees plaintext.
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

	record := models.SecretRecord{
		Name:             name,
		Type:             "account",
		Salt:             salt,
		Nonce:            nonce,
		EncryptedPayload: ciphertext,
	}

	if err := database.Insert(db, record); err != nil {
		// Surface a friendly duplicate-name error without leaking internal details.
		return fmt.Errorf("[X] Error: a credential named %q already exists. Use `vext update` to modify it", name)
	}

	fmt.Printf("[✓] Credential %q saved.\n", name)
	return nil
}
