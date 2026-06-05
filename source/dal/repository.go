package dal

import (
	"context"

	"vextpss/source/core"
)

// SecretRepository defines the persistence contract for secrets.
// Implementations live in this package — never in core or application.
type SecretRepository interface {
	// Save persists a secret with its pre-encrypted payload. Plaintext never crosses this boundary.
	Save(ctx context.Context, secret *core.Secret, encrypted []byte) error

	// GetByName retrieves a secret by name. Returns the domain entity and the encrypted payload.
	// Returns core.ErrSecretNotFound if no record exists.
	GetByName(ctx context.Context, name string) (*core.Secret, []byte, error)

	// ListAll returns metadata for all secrets without decrypting payloads.
	ListAll(ctx context.Context) ([]core.Secret, error)

	// Delete removes a secret by name. Returns core.ErrSecretNotFound if it does not exist.
	Delete(ctx context.Context, name string) error
}
