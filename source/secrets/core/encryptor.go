package core

import (
	"context"
	"errors"
)

var (
	ErrDecryptionFailed     = errors.New("wrong password or corrupted data")
	ErrCorruptPayload       = errors.New("payload is corrupt or truncated")
	ErrUnsupportedAlgorithm = errors.New("unsupported encryption algorithm")
)

type Encrypted struct {
	Algorithm  string
	Ciphertext []byte
	Metadata   []byte
}

type Encryptor interface {
	Algorithm() string
	Encrypt(ctx context.Context, plaintext, password []byte) (Encrypted, error)
	Decrypt(ctx context.Context, payload Encrypted, password []byte) ([]byte, error)
}
