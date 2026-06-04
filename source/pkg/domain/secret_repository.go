package domain

import "context"

// SecretRepository defines the persistence contract for secrets.
// Implementations live in infrastructure/storage — never in domain or application.
type SecretRepository interface {
	// Save persists a secret with its pre-encrypted payload. Plaintext never crosses this boundary.
	Save(ctx context.Context, secret *Secret, encryptedPayload []byte) error

	// GetByName retrieves a secret by name. Returns the domain entity and the encrypted payload.
	// Returns ErrSecretNotFound if no record exists.
	GetByName(ctx context.Context, name string) (*Secret, []byte, error)

	// ListAll returns metadata for all secrets without decrypting payloads.
	ListAll(ctx context.Context) ([]Secret, error)

	// Delete removes a secret by name. Returns ErrSecretNotFound if it does not exist.
	Delete(ctx context.Context, name string) error
}
