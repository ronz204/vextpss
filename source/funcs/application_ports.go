package funcs

import (
	"context"

	"vextpss/source/secrets"
)

// SecretRepo is the persistence contract required by use cases.
type Repository interface {
	Create(ctx context.Context, secret *secrets.Secret, encrypted []byte) error
	GetByName(ctx context.Context, name string) (*secrets.Secret, []byte, error)
	Update(ctx context.Context, secret *secrets.Secret, encrypted []byte) error
	Delete(ctx context.Context, name string) error
	ListAll(ctx context.Context) ([]secrets.Secret, error)
	GetAll(ctx context.Context) ([]secrets.Credential, error)
}

// Encryptor is the encryption contract required by use cases.
type Encryptor interface {
	Encrypt(ctx context.Context, plaintext, password []byte) (salt, nonce, ciphertext []byte, err error)
	Decrypt(ctx context.Context, password, salt, nonce, ciphertext []byte) ([]byte, error)
}

// InputCollector is the user-input contract required by adapters.
type Collector interface {
	Payload(secretType string) ([]byte, error)
	Confirm(prompt string) (bool, error)
	Master() ([]byte, error)
}
