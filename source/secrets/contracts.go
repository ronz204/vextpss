package secrets

import "context"

type Repository interface {
	Delete(ctx context.Context, name string) error
	Create(ctx context.Context, secret Secret) error
	Update(ctx context.Context, secret Secret) error
	ListAll(ctx context.Context) ([]Secret, error)
}

type Encryptor interface {
	Encrypt(ctx context.Context, plaintext, password []byte) (salt, nonce, ciphertext []byte, err error)
	Decrypt(ctx context.Context, salt, nonce, ciphertext []byte, password []byte) (plaintext []byte, err error)
}
