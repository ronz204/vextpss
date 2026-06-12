package core

import "context"

type EncInDto struct {
	Plaintext []byte
	Password  []byte
}

type EncOutDto struct {
	Salt       []byte
	Nonce      []byte
	Ciphertext []byte
}

type DecInDto struct {
	Password   []byte
	Salt       []byte
	Nonce      []byte
	Ciphertext []byte
}

// Encryptor is the cryptographic contract consumed by use cases.
type Encryptor interface {
	// Encrypt ciphers plaintext with password.
	Encrypt(ctx context.Context, in EncInDto) (out EncOutDto, err error)
	// Decrypt reverses encrypt with password.
	Decrypt(ctx context.Context, in DecInDto) ([]byte, error)
}
