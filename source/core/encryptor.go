package core

import "context"

type EncryptInDto struct {
	Plaintext []byte
	Password  []byte
}

type EncryptOutDto struct {
	Salt       []byte
	Nonce      []byte
	Ciphertext []byte
}

type DecryptInDto struct {
	Password   []byte
	Salt       []byte
	Nonce      []byte
	Ciphertext []byte
}

// Encryptor is the cryptographic contract consumed by use cases.
type Encryptor interface {
	// Encrypt ciphers plaintext with password.
	Encrypt(ctx context.Context, inDto EncryptInDto) (outDto EncryptOutDto, err error)
	// Decrypt reverses encrypt with password.
	Decrypt(ctx context.Context, inDto DecryptInDto) ([]byte, error)
}
