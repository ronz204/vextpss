package secrets

import (
	"context"
)

type Repository interface {
	Create(ctx context.Context, secret *Secret, encrypted []byte) error
	GetByName(ctx context.Context, name string) (*Secret, []byte, error)
	Update(ctx context.Context, secret *Secret, encrypted []byte) error

	Delete(ctx context.Context, name string) error
	ListAll(ctx context.Context) ([]Secret, error)
	GetAll(ctx context.Context) ([]Credential, error)
}

type Encryptor interface {
	Encrypt(ctx context.Context, plaintext, password []byte) (salt, nonce, ciphertext []byte, err error)
	Decrypt(ctx context.Context, password, salt, nonce, ciphertext []byte) ([]byte, error)
}
