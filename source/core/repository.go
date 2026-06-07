package core

import "context"

// FullRecord pairs a Secret with its raw encrypted payload for bulk operations (e.g. export).
type FullRecord struct {
	Secret    Secret
	Encrypted []byte
}

// SecretRepository is the persistence contract consumed by use cases.
// Implementations live in dal/repos/ — never in core or application.
type SecretRepository interface {
	// Save persists a new secret with its pre-encrypted payload.
	Save(ctx context.Context, secret *Secret, encrypted []byte) error

	// GetByName retrieves a secret by name. Returns core.ErrSecretNotFound if absent.
	GetByName(ctx context.Context, name string) (*Secret, []byte, error)

	// ListAll returns metadata for all secrets without decrypting payloads.
	ListAll(ctx context.Context) ([]Secret, error)

	// GetAll returns all secrets together with their encrypted payloads.
	GetAll(ctx context.Context) ([]FullRecord, error)

	// Delete removes a secret by name. Returns core.ErrSecretNotFound if absent.
	Delete(ctx context.Context, name string) error

	// Update replaces the encrypted payload and crypto material for an existing secret.
	Update(ctx context.Context, secret *Secret, encrypted []byte) error
}
