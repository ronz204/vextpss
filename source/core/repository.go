package core

import (
	"context"
	"vextpss/source/secrets"
)

// SecretRepository is the persistence contract consumed by use cases.
type SecretRepository interface {
	// Delete removes a secret by name. Returns core.ErrSecretNotFound if absent.
	Delete(ctx context.Context, name string) error

	// ListAll returns metadata for all secrets without decrypting payloads.
	ListAll(ctx context.Context) ([]secrets.Secret, error)

	// GetAll returns all secrets together with their encrypted payloads.
	GetAll(ctx context.Context) ([]secrets.Credential, error)

	// Save persists a new secret with its pre-encrypted payload.
	Create(ctx context.Context, secret *secrets.Secret, encrypted []byte) error

	// GetByName retrieves a secret by name. Returns core.ErrSecretNotFound if absent.
	GetByName(ctx context.Context, name string) (*secrets.Secret, []byte, error)

	// Update replaces the encrypted payload and crypto material for an existing secret.
	Update(ctx context.Context, secret *secrets.Secret, encrypted []byte) error
}
